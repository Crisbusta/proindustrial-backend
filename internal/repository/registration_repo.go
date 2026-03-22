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

func (r *RegistrationRepo) Create(in CreateRegistrationInput) (*model.ProviderRegistration, error) {
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
		pq.StringArray(in.Services),
		nullableStr(in.Description),
	).StructScan(&reg)
	return &reg, err
}
