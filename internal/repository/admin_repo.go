package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/crisbusta/proindustrial-backend-public/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var nonSlugCharPattern = regexp.MustCompile(`[^a-z0-9]+`)

type AdminRepo struct {
	db *sqlx.DB
}

func NewAdminRepo(db *sqlx.DB) *AdminRepo {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) ListRegistrations(status string) ([]model.ProviderRegistration, error) {
	regs := []model.ProviderRegistration{}
	query := `SELECT * FROM provider_registrations`
	args := []interface{}{}
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC`
	err := r.db.Select(&regs, query, args...)
	return regs, err
}

func (r *AdminRepo) GetRegistrationByID(id string) (*model.ProviderRegistration, error) {
	var reg model.ProviderRegistration
	err := r.db.Get(&reg, `SELECT * FROM provider_registrations WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

type ApproveRegistrationResult struct {
	Registration    model.ProviderRegistration `json:"registration"`
	Company         model.Company              `json:"company"`
	User            model.User                 `json:"user"`
	InitialPassword string                     `json:"initialPassword"`
	EmailStatus     string                     `json:"emailStatus,omitempty"`
	EmailNote       string                     `json:"emailNote,omitempty"`
}

func (r *AdminRepo) ApproveRegistration(id, passwordHash, initialPassword string) (*ApproveRegistrationResult, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var reg model.ProviderRegistration
	err = tx.Get(&reg, `SELECT * FROM provider_registrations WHERE id = $1 FOR UPDATE`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRegistrationNotFound
		}
		// Un fallo de conexión o de bloqueo no es un "no encontrado":
		// devolverlo como 404 mandaba al admin a buscar un registro que sí existe.
		return nil, err
	}
	if reg.Status != "pending" {
		return nil, ErrRegistrationAlreadyDone
	}

	email := strings.ToLower(strings.TrimSpace(reg.Email))

	var existingCount int
	if err := tx.Get(&existingCount, `SELECT COUNT(*) FROM users WHERE lower(email) = $1`, email); err != nil {
		return nil, err
	}
	if existingCount > 0 {
		return nil, ErrRegistrationEmailInUse
	}

	categories := pq.StringArray(reg.Services)
	serviceLabels := pq.StringArray(categoryNames(reg.Services))
	slug, err := nextCompanySlug(tx, reg.CompanyName)
	if err != nil {
		return nil, err
	}

	var company model.Company
	err = tx.QueryRowx(`
		INSERT INTO companies (
			slug, name, description, location, region, categories, services, phone, email, featured
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, false)
		RETURNING `+companyCols,
		slug,
		reg.CompanyName,
		truncatedDescription(reg.Description),
		reg.Region,
		reg.Region,
		categories,
		serviceLabels,
		reg.Phone,
		email,
	).StructScan(&company)
	if err != nil {
		return nil, err
	}

	var user model.User
	err = tx.QueryRowx(`
		INSERT INTO users (email, password_hash, company_id, role, must_change_password)
		VALUES ($1, $2, $3, 'provider', true)
		RETURNING *`,
		email,
		passwordHash,
		company.ID,
	).StructScan(&user)
	if err != nil {
		return nil, err
	}

	err = tx.Get(&reg, `
		UPDATE provider_registrations
		SET status = 'approved', company_id = $2, user_id = $3, approved_at = $4, email = $5
		WHERE id = $1
		RETURNING *`,
		id, company.ID, user.ID, time.Now(), email,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &ApproveRegistrationResult{
		Registration:    reg,
		Company:         company,
		User:            user,
		InitialPassword: initialPassword,
	}, nil
}

// RejectRegistration marca el registro como rechazado y guarda el motivo.
// El motivo es opcional; queda registrado para que la decisión sea auditable.
func (r *AdminRepo) RejectRegistration(id, reason string) (*model.ProviderRegistration, error) {
	var reg model.ProviderRegistration
	err := r.db.Get(&reg, `
		UPDATE provider_registrations
		SET status = 'rejected', rejection_reason = $2, rejected_at = $3
		WHERE id = $1 AND status = 'pending'
		RETURNING *`,
		id, nullableStr(reason), time.Now(),
	)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		// Cero filas: o el id no existe, o el registro ya no está pendiente.
		checkErr := r.db.Get(&reg, `SELECT * FROM provider_registrations WHERE id = $1`, id)
		if checkErr != nil {
			if errors.Is(checkErr, sql.ErrNoRows) {
				return nil, ErrRegistrationNotFound
			}
			return nil, checkErr
		}
		return nil, ErrRegistrationAlreadyDone
	}
	return &reg, nil
}

func (r *AdminRepo) DeleteApprovedCompanyByRegistration(id string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var reg model.ProviderRegistration
	if err := tx.Get(&reg, `SELECT * FROM provider_registrations WHERE id = $1 FOR UPDATE`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRegistrationNotFound
		}
		return err
	}
	if reg.Status != "approved" {
		return ErrApprovedCompanyNotFound
	}

	// El vínculo autoritativo es company_id, que la aprobación guarda.
	// Buscar por correo podía quedarse sin resultado (si el proveedor
	// cambió su email en el panel) o devolver la empresa equivocada
	// (companies.email no es único). El camino por correo se conserva
	// solo para registros aprobados antes de la migración 011, y exige
	// que la coincidencia sea inequívoca.
	companyID := ""
	if reg.CompanyID.Valid {
		companyID = reg.CompanyID.String
	} else {
		var matches []string
		if err := tx.Select(&matches, `SELECT id FROM companies WHERE lower(email) = lower($1)`, reg.Email); err != nil {
			return err
		}
		if len(matches) != 1 {
			return ErrApprovedCompanyNotFound
		}
		companyID = matches[0]
	}

	var exists int
	if err := tx.Get(&exists, `SELECT COUNT(*) FROM companies WHERE id = $1`, companyID); err != nil {
		return err
	}
	if exists == 0 {
		return ErrApprovedCompanyNotFound
	}

	if _, err := tx.Exec(`DELETE FROM quote_requests WHERE target_company_id = $1`, companyID); err != nil {
		return err
	}
	// Por company_id, no por correo: así no se arrastra a un usuario
	// ajeno que casualmente comparta dirección.
	if _, err := tx.Exec(`DELETE FROM users WHERE company_id = $1`, companyID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM companies WHERE id = $1`, companyID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM provider_registrations WHERE id = $1`, id); err != nil {
		return err
	}

	return tx.Commit()
}

func nextCompanySlug(tx *sqlx.Tx, companyName string) (string, error) {
	base := slugify(companyName)
	if base == "" {
		base = "empresa"
	}

	slug := base
	for i := 2; ; i++ {
		var count int
		if err := tx.Get(&count, `SELECT COUNT(*) FROM companies WHERE slug = $1`, slug); err != nil {
			return "", err
		}
		if count == 0 {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}

// slugReplacer translitera los acentos del castellano antes de filtrar.
// Sin esto, "Tuberías del Sur" quedaba como "tuber-as-del-sur".
var slugReplacer = strings.NewReplacer(
	"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n",
	"Á", "a", "É", "e", "Í", "i", "Ó", "o", "Ú", "u", "Ü", "u", "Ñ", "n",
	"à", "a", "è", "e", "ì", "i", "ò", "o", "ù", "u", "ç", "c",
)

func slugify(value string) string {
	slug := strings.ToLower(strings.TrimSpace(slugReplacer.Replace(value)))
	slug = nonSlugCharPattern.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	// companies.slug es VARCHAR(100): sin truncar, un nombre largo
	// hacía fallar el INSERT con un 500 opaco justo al aprobar.
	if len(slug) > model.MaxSlugBase {
		slug = strings.Trim(slug[:model.MaxSlugBase], "-")
	}
	return slug
}

// truncatedDescription recorta al tope canónico. Los registros anteriores
// a la migración 011 pueden exceder el límite y, sin recorte, el INSERT en
// companies fallaba contra el CHECK o contra el tope de 1 MB del tsvector
// generado, dejando esos registros imposibles de aprobar.
func truncatedDescription(d model.NullString) model.NullString {
	if !d.Valid {
		return d
	}
	runes := []rune(d.String)
	if len(runes) <= model.MaxDescription {
		return d
	}
	d.String = strings.TrimSpace(string(runes[:model.MaxDescription]))
	return d
}

// SetEmailStatus registra el resultado del envío del correo de acceso.
// Se persiste para que el admin pueda ver después si llegó o no, en vez
// de depender de un banner que desaparece al recargar.
func (r *AdminRepo) SetEmailStatus(registrationID, status, note string) error {
	_, err := r.db.Exec(`
		UPDATE provider_registrations
		SET email_status = $2, email_note = $3
		WHERE id = $1`, registrationID, status, note)
	return err
}

// ResetInitialPassword genera un acceso nuevo para un registro ya aprobado.
// Es la salida cuando el correo no llegó o la contraseña inicial se perdió.
func (r *AdminRepo) ResetInitialPassword(registrationID, passwordHash string) (*ApproveRegistrationResult, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var reg model.ProviderRegistration
	if err := tx.Get(&reg, `SELECT * FROM provider_registrations WHERE id = $1 FOR UPDATE`, registrationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRegistrationNotFound
		}
		return nil, err
	}
	if reg.Status != "approved" {
		return nil, ErrApprovedCompanyNotFound
	}

	var user model.User
	query := `UPDATE users SET password_hash = $1, must_change_password = true WHERE id = $2 RETURNING *`
	arg := reg.UserID.String
	if !reg.UserID.Valid {
		// Registros aprobados antes de la migración 011.
		query = `UPDATE users SET password_hash = $1, must_change_password = true WHERE lower(email) = lower($2) RETURNING *`
		arg = reg.Email
	}
	if err := tx.QueryRowx(query, passwordHash, arg).StructScan(&user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrApprovedCompanyNotFound
		}
		return nil, err
	}

	var company model.Company
	if err := tx.Get(&company, `SELECT `+companyCols+` FROM companies WHERE id = $1`, user.CompanyID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &ApproveRegistrationResult{
		Registration: reg,
		Company:      company,
		User:         user,
	}, nil
}

func categoryNames(slugs []string) []string {
	if len(slugs) == 0 {
		return []string{}
	}

	names := make([]string, 0, len(slugs))
	for _, slug := range slugs {
		label := slug
		for _, group := range CategoryGroups {
			if group.Slug == slug {
				label = group.Name
				break
			}
		}
		names = append(names, label)
	}
	return names
}
