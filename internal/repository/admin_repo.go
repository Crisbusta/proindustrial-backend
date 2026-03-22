package repository

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/crisbusta/proindustrial-backend-public/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrRegistrationNotFound    = errors.New("registration not found")
	ErrRegistrationAlreadyDone = errors.New("registration already processed")
	ErrRegistrationEmailInUse  = errors.New("registration email already in use")
	ErrApprovedCompanyNotFound = errors.New("approved company not found")
	nonSlugCharPattern         = regexp.MustCompile(`[^a-z0-9]+`)
)

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
		return nil, ErrRegistrationNotFound
	}
	if reg.Status != "pending" {
		return nil, ErrRegistrationAlreadyDone
	}

	var existingCount int
	if err := tx.Get(&existingCount, `SELECT COUNT(*) FROM users WHERE email = $1`, reg.Email); err != nil {
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
		RETURNING *`,
		slug,
		reg.CompanyName,
		reg.Description,
		reg.Region,
		reg.Region,
		categories,
		serviceLabels,
		reg.Phone,
		reg.Email,
	).StructScan(&company)
	if err != nil {
		return nil, err
	}

	var user model.User
	err = tx.QueryRowx(`
		INSERT INTO users (email, password_hash, company_id, role, must_change_password)
		VALUES ($1, $2, $3, 'provider', true)
		RETURNING *`,
		reg.Email,
		passwordHash,
		company.ID,
	).StructScan(&user)
	if err != nil {
		return nil, err
	}

	err = tx.Get(&reg, `
		UPDATE provider_registrations
		SET status = 'approved'
		WHERE id = $1
		RETURNING *`,
		id,
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

func (r *AdminRepo) RejectRegistration(id string) (*model.ProviderRegistration, error) {
	var reg model.ProviderRegistration
	err := r.db.Get(&reg, `
		UPDATE provider_registrations
		SET status = 'rejected'
		WHERE id = $1 AND status = 'pending'
		RETURNING *`,
		id,
	)
	if err != nil {
		checkErr := r.db.Get(&reg, `SELECT * FROM provider_registrations WHERE id = $1`, id)
		if checkErr != nil {
			return nil, ErrRegistrationNotFound
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
		return ErrRegistrationNotFound
	}
	if reg.Status != "approved" {
		return ErrApprovedCompanyNotFound
	}

	var companyID string
	if err := tx.Get(&companyID, `SELECT id FROM companies WHERE email = $1`, reg.Email); err != nil {
		return ErrApprovedCompanyNotFound
	}

	if _, err := tx.Exec(`DELETE FROM quote_requests WHERE target_company_id = $1`, companyID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM users WHERE email = $1`, reg.Email); err != nil {
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

func slugify(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	slug = nonSlugCharPattern.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
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
