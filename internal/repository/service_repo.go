package repository

import (
	"database/sql"
	"fmt"

	"github.com/crisbusta/proindustrial-backend-public/internal/model"
	"github.com/jmoiron/sqlx"
)

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

type ServiceRepo struct {
	db *sqlx.DB
}

func NewServiceRepo(db *sqlx.DB) *ServiceRepo {
	return &ServiceRepo{db: db}
}

type CreateServiceInput struct {
	Name        string
	Category    string
	Description string
}

func (r *ServiceRepo) List(companyID string) ([]model.CompanyService, error) {
	services := []model.CompanyService{}
	err := r.db.Select(&services,
		`SELECT * FROM company_services WHERE company_id = $1 ORDER BY created_at DESC`,
		companyID,
	)
	return services, err
}

func (r *ServiceRepo) ListActive(companyID string) ([]model.CompanyService, error) {
	services := []model.CompanyService{}
	err := r.db.Select(&services,
		`SELECT * FROM company_services WHERE company_id = $1 AND status = 'active' ORDER BY created_at ASC`,
		companyID,
	)
	return services, err
}

func (r *ServiceRepo) Create(companyID string, in CreateServiceInput) (*model.CompanyService, error) {
	var s model.CompanyService
	err := r.db.QueryRowx(`
		INSERT INTO company_services (company_id, name, category, description)
		VALUES ($1,$2,$3,$4)
		RETURNING *`,
		companyID,
		in.Name,
		model.NullString{NullString: toNullString(in.Category)},
		model.NullString{NullString: toNullString(in.Description)},
	).StructScan(&s)
	return &s, err
}

func (r *ServiceRepo) Update(id, companyID string, fields map[string]interface{}) (*model.CompanyService, error) {
	sets := []string{}
	args := []interface{}{}
	idx := 1
	for k, v := range fields {
		sets = append(sets, fmt.Sprintf("%s = $%d", k, idx))
		args = append(args, v)
		idx++
	}
	args = append(args, id, companyID)
	query := fmt.Sprintf(
		"UPDATE company_services SET %s WHERE id = $%d AND company_id = $%d RETURNING *",
		joinSets(sets), idx, idx+1,
	)
	var s model.CompanyService
	err := r.db.QueryRowx(query, args...).StructScan(&s)
	return &s, err
}

func (r *ServiceRepo) Delete(id, companyID string) error {
	res, err := r.db.Exec(
		`DELETE FROM company_services WHERE id = $1 AND company_id = $2`,
		id, companyID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func (r *ServiceRepo) Count(companyID string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM company_services WHERE company_id = $1`, companyID).Scan(&count)
	return count, err
}

func joinSets(sets []string) string {
	result := ""
	for i, s := range sets {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
