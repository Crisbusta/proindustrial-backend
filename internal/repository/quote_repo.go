package repository

import (
	"fmt"

	"github.com/crisbusta/proindustrial-backend-public/internal/model"
	"github.com/jmoiron/sqlx"
)

type QuoteRepo struct {
	db *sqlx.DB
}

func NewQuoteRepo(db *sqlx.DB) *QuoteRepo {
	return &QuoteRepo{db: db}
}

type CreateQuoteInput struct {
	RequesterName    string
	RequesterCompany string
	RequesterEmail   string
	RequesterPhone   string
	Service          string
	Description      string
	Location         string
	TargetCompanyID  string
}

func nullableStr(s string) model.NullString {
	ns := model.NullString{}
	if s != "" {
		ns.Valid = true
		ns.String = s
	}
	return ns
}

func (r *QuoteRepo) Create(in CreateQuoteInput) (*model.QuoteRequest, error) {
	var q model.QuoteRequest
	err := r.db.QueryRowx(`
		INSERT INTO quote_requests
			(requester_name, requester_company, requester_email, requester_phone,
			 service, description, location, target_company_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING *`,
		in.RequesterName, nullableStr(in.RequesterCompany), in.RequesterEmail,
		nullableStr(in.RequesterPhone), in.Service, nullableStr(in.Description),
		nullableStr(in.Location), nullableStr(in.TargetCompanyID),
	).StructScan(&q)
	return &q, err
}

func (r *QuoteRepo) ListByCompany(companyID, status string) ([]model.QuoteRequest, error) {
	query := `SELECT * FROM quote_requests WHERE target_company_id = $1`
	args := []interface{}{companyID}
	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"
	quotes := []model.QuoteRequest{}
	err := r.db.Select(&quotes, query, args...)
	return quotes, err
}

func (r *QuoteRepo) UpdateStatus(id, companyID, status string) error {
	res, err := r.db.Exec(
		`UPDATE quote_requests SET status = $1 WHERE id = $2 AND target_company_id = $3`,
		status, id, companyID,
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

// DashboardStats holds KPIs for the panel dashboard
type DashboardStats struct {
	TotalQuotes     int `json:"totalQuotes"`
	NewQuotes       int `json:"newQuotes"`
	ReadQuotes      int `json:"readQuotes"`
	RespondedQuotes int `json:"respondedQuotes"`
	TotalServices   int `json:"totalServices"`
}

func (r *QuoteRepo) Stats(companyID string) (*DashboardStats, error) {
	var s DashboardStats
	err := r.db.QueryRow(`
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status='new'),
			COUNT(*) FILTER (WHERE status='read'),
			COUNT(*) FILTER (WHERE status='responded')
		FROM quote_requests WHERE target_company_id = $1`,
		companyID,
	).Scan(&s.TotalQuotes, &s.NewQuotes, &s.ReadQuotes, &s.RespondedQuotes)
	return &s, err
}
