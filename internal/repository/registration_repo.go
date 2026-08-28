package repository

import (
	"github.com/crisbusta/proindustrial-backend-public/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type RegistrationRepo struct {
	db *sqlx.DB
}

func NewRegistrationRepo(db *sqlx.DB) *RegistrationRepo {
	return &RegistrationRepo{db: db}
}

type CreateRegistrationInput struct {
	CompanyName string
	Email       string
	Phone       string
	Region      string
	Services    []string
	Description string
}

// PendingExists indica si ya hay una solicitud en revisión para ese correo.
// El correo debe venir normalizado en minúsculas.
func (r *RegistrationRepo) PendingExists(email string) (bool, error) {
	var n int
	err := r.db.Get(&n, `
		SELECT COUNT(*) FROM provider_registrations
		WHERE lower(email) = $1 AND status = 'pending'`, email)
	return n > 0, err
}

func (r *RegistrationRepo) Create(in CreateRegistrationInput) (*model.ProviderRegistration, error) {
	// Un slice nil se inserta como NULL explícito y anula el DEFAULT '{}'
	// de la columna. Al leerlo, el JSON sale como "services": null y el
	// panel del admin —que asume un array— se cae entero.
	services := in.Services
	if services == nil {
		services = []string{}
	}

	var reg model.ProviderRegistration
	err := r.db.QueryRowx(`
		INSERT INTO provider_registrations
			(company_name, email, phone, region, services, description)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING *`,
		in.CompanyName,
		in.Email,
		nullableStr(in.Phone),
		nullableStr(in.Region),
		pq.StringArray(services),
		nullableStr(in.Description),
	).StructScan(&reg)
	return &reg, err
}
