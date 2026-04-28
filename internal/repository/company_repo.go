package repository

import (
	"fmt"
	"strings"

	"github.com/crisbusta/proindustrial-backend-public/internal/model"
	"github.com/jmoiron/sqlx"
)

type CompanyRepo struct {
	db *sqlx.DB
}

func NewCompanyRepo(db *sqlx.DB) *CompanyRepo {
	return &CompanyRepo{db: db}
}

const companyCols = `id, slug, name, tagline, description, location, region, categories, services, phone, email, website, years_active, featured, logo_url, cover_url, created_at, updated_at`

func (r *CompanyRepo) List(category, region, q string, featured *bool) ([]model.Company, error) {
	query := `SELECT ` + companyCols + ` FROM companies WHERE 1=1`
	args := []interface{}{}
	idx := 1
	qIdx := 0

	if category != "" {
		query += fmt.Sprintf(" AND $%d = ANY(categories)", idx)
		args = append(args, category)
		idx++
	}
	if region != "" {
		query += fmt.Sprintf(" AND region = $%d", idx)
		args = append(args, region)
		idx++
	}
	if featured != nil {
		query += fmt.Sprintf(" AND featured = $%d", idx)
		args = append(args, *featured)
		idx++
	}
	if q != "" {
		qIdx = idx
		query += fmt.Sprintf(" AND tsv @@ plainto_tsquery('spanish', $%d)", idx)
		args = append(args, q)
		idx++
	}
	_ = strings.TrimSpace(query)
	_ = idx

	if qIdx > 0 {
		query += fmt.Sprintf(" ORDER BY ts_rank(tsv, plainto_tsquery('spanish', $%d)) DESC, featured DESC, name ASC", qIdx)
	} else {
		query += " ORDER BY featured DESC, name ASC"
	}

	companies := []model.Company{}
	err := r.db.Select(&companies, query, args...)
	return companies, err
}

func (r *CompanyRepo) GetBySlug(slug string) (*model.Company, error) {
	var company model.Company
	err := r.db.Get(&company, "SELECT "+companyCols+" FROM companies WHERE slug = $1", slug)
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *CompanyRepo) GetByID(id string) (*model.Company, error) {
	var company model.Company
	err := r.db.Get(&company, "SELECT "+companyCols+" FROM companies WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *CompanyRepo) Update(id string, fields map[string]interface{}) error {
	sets := []string{}
	args := []interface{}{}
	idx := 1
	for k, v := range fields {
		sets = append(sets, fmt.Sprintf("%s = $%d", k, idx))
		args = append(args, v)
		idx++
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE companies SET %s, updated_at = NOW() WHERE id = $%d", strings.Join(sets, ", "), idx)
	_, err := r.db.Exec(query, args...)
	return err
}
