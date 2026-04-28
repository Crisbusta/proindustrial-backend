package repository

import (
	"sort"
	"time"

	"github.com/crisbusta/proindustrial-backend-public/internal/model"
	"github.com/jmoiron/sqlx"
)

type EventRepo struct {
	db *sqlx.DB
}

func NewEventRepo(db *sqlx.DB) *EventRepo {
	return &EventRepo{db: db}
}

func (r *EventRepo) Insert(companyID, eventType, visitorID, referrer, ipHash string) error {
	_, err := r.db.Exec(`
		INSERT INTO company_events (company_id, event_type, visitor_id, referrer, ip_hash)
		VALUES ($1, $2, NULLIF($3,''), NULLIF($4,''), NULLIF($5,''))`,
		companyID, eventType, visitorID, referrer, ipHash,
	)
	return err
}

func (r *EventRepo) GetAnalytics(companyID string, days int) (*model.AnalyticsResult, error) {
	// ── Totals ────────────────────────────────────────────────────────────────
	var totals model.AnalyticsTotals
	err := r.db.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE event_type = 'profile_view'),
			COUNT(*) FILTER (WHERE event_type IN ('contact_click_phone','contact_click_whatsapp','contact_click_email')),
			COUNT(*) FILTER (WHERE event_type = 'quote_form_open'),
			COUNT(*) FILTER (WHERE event_type = 'quote_form_submit')
		FROM company_events
		WHERE company_id = $1
		  AND created_at >= NOW() - ($2::int * INTERVAL '1 day')`,
		companyID, days,
	).Scan(&totals.ProfileViews, &totals.ContactClicks, &totals.QuoteFormOpens, &totals.QuoteFormSubmits)
	if err != nil {
		return nil, err
	}

	_ = r.db.QueryRow(`
		SELECT COUNT(*) FROM quote_requests
		WHERE target_company_id = $1
		  AND created_at >= NOW() - ($2::int * INTERVAL '1 day')`,
		companyID, days,
	).Scan(&totals.RFQsReceived)

	if totals.ProfileViews > 0 {
		totals.ContactRate = float64(totals.ContactClicks) / float64(totals.ProfileViews)
	}

	// ── Trend (daily) ─────────────────────────────────────────────────────────
	type row struct {
		Day     time.Time `db:"day"`
		PV      int       `db:"pv"`
		CC      int       `db:"cc"`
	}
	var rows []row
	err = r.db.Select(&rows, `
		SELECT
			DATE(created_at AT TIME ZONE 'UTC') AS day,
			COUNT(*) FILTER (WHERE event_type = 'profile_view')                                                               AS pv,
			COUNT(*) FILTER (WHERE event_type IN ('contact_click_phone','contact_click_whatsapp','contact_click_email')) AS cc
		FROM company_events
		WHERE company_id = $1
		  AND created_at >= NOW() - ($2::int * INTERVAL '1 day')
		GROUP BY day
		ORDER BY day ASC`,
		companyID, days,
	)
	if err != nil {
		rows = nil
	}

	// RFQs per day
	type rfqRow struct {
		Day  time.Time `db:"day"`
		Cnt  int       `db:"cnt"`
	}
	var rfqRows []rfqRow
	_ = r.db.Select(&rfqRows, `
		SELECT DATE(created_at AT TIME ZONE 'UTC') AS day, COUNT(*) AS cnt
		FROM quote_requests
		WHERE target_company_id = $1
		  AND created_at >= NOW() - ($2::int * INTERVAL '1 day')
		GROUP BY day`,
		companyID, days,
	)
	rfqByDay := map[string]int{}
	for _, rr := range rfqRows {
		rfqByDay[rr.Day.Format("2006-01-02")] = rr.Cnt
	}

	// Build complete day-by-day series (fill gaps with 0)
	evByDay := map[string]row{}
	for _, r := range rows {
		evByDay[r.Day.Format("2006-01-02")] = r
	}

	now := time.Now().UTC()
	trend := make([]model.DailyMetric, 0, days)
	for i := days - 1; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		ev := evByDay[d]
		trend = append(trend, model.DailyMetric{
			Date:          d,
			ProfileViews:  ev.PV,
			ContactClicks: ev.CC,
			RFQs:          rfqByDay[d],
		})
	}
	sort.Slice(trend, func(i, j int) bool { return trend[i].Date < trend[j].Date })

	rangeLabel := map[int]string{7: "7d", 30: "30d", 90: "90d"}[days]
	if rangeLabel == "" {
		rangeLabel = "30d"
	}

	return &model.AnalyticsResult{
		Range:  rangeLabel,
		Days:   days,
		Totals: totals,
		Trend:  trend,
	}, nil
}
