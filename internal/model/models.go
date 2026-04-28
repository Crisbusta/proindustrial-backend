package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

// NullString wraps sql.NullString with proper JSON marshaling
type NullString struct {
	sql.NullString
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

func (ns *NullString) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		ns.Valid = false
		return nil
	}
	ns.Valid = true
	return json.Unmarshal(b, &ns.String)
}

func (ns NullString) Value() (driver.Value, error) {
	return ns.NullString.Value()
}

func (ns *NullString) Scan(value interface{}) error {
	return ns.NullString.Scan(value)
}

// NullInt64 wraps sql.NullInt64 with proper JSON marshaling
type NullInt64 struct {
	sql.NullInt64
}

func (ni NullInt64) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ni.Int64)
}

func (ni *NullInt64) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		ni.Valid = false
		return nil
	}
	ni.Valid = true
	return json.Unmarshal(b, &ni.Int64)
}

func (ni NullInt64) Value() (driver.Value, error) {
	return ni.NullInt64.Value()
}

func (ni *NullInt64) Scan(value interface{}) error {
	return ni.NullInt64.Scan(value)
}

type Company struct {
	ID          string         `db:"id"          json:"id"`
	Slug        string         `db:"slug"        json:"slug"`
	Name        string         `db:"name"        json:"name"`
	Tagline     NullString     `db:"tagline"     json:"tagline"`
	Description NullString     `db:"description" json:"description"`
	Location    NullString     `db:"location"    json:"location"`
	Region      NullString     `db:"region"      json:"region"`
	Categories  pq.StringArray `db:"categories"  json:"categories"`
	Services    pq.StringArray `db:"services"    json:"services"`
	Phone       NullString     `db:"phone"       json:"phone"`
	Email       NullString     `db:"email"       json:"email"`
	Website     NullString     `db:"website"     json:"website"`
	YearsActive NullInt64      `db:"years_active" json:"yearsActive"`
	Featured    bool           `db:"featured"    json:"featured"`
	LogoURL     NullString     `db:"logo_url"    json:"logoUrl"`
	CoverURL    NullString     `db:"cover_url"   json:"coverUrl"`
	CreatedAt   time.Time      `db:"created_at"  json:"createdAt"`
	UpdatedAt   time.Time      `db:"updated_at"  json:"updatedAt"`
}

type ServiceImage struct {
	ID        string     `db:"id"         json:"id"`
	ServiceID string     `db:"service_id" json:"serviceId"`
	URL       string     `db:"url"        json:"url"`
	AltText   NullString `db:"alt_text"   json:"altText"`
	SortOrder int        `db:"sort_order" json:"sortOrder"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
}

type CompanyCertification struct {
	ID          string     `db:"id"           json:"id"`
	CompanyID   string     `db:"company_id"   json:"companyId"`
	Name        string     `db:"name"         json:"name"`
	Issuer      NullString `db:"issuer"       json:"issuer"`
	DocumentURL NullString `db:"document_url" json:"documentUrl"`
	IssuedAt    NullString `db:"issued_at"    json:"issuedAt"`
	ExpiresAt   NullString `db:"expires_at"   json:"expiresAt"`
	CreatedAt   time.Time  `db:"created_at"   json:"createdAt"`
}

type CompanyProject struct {
	ID          string         `db:"id"          json:"id"`
	CompanyID   string         `db:"company_id"  json:"companyId"`
	Title       string         `db:"title"       json:"title"`
	Description NullString     `db:"description" json:"description"`
	ClientName  NullString     `db:"client_name" json:"clientName"`
	Year        NullInt64      `db:"year"        json:"year"`
	CoverURL    NullString     `db:"cover_url"   json:"coverUrl"`
	SortOrder   int            `db:"sort_order"  json:"sortOrder"`
	CreatedAt   time.Time      `db:"created_at"  json:"createdAt"`
	Images      []ProjectImage `db:"-"           json:"images"`
}

type ProjectImage struct {
	ID        string     `db:"id"         json:"id"`
	ProjectID string     `db:"project_id" json:"projectId"`
	URL       string     `db:"url"        json:"url"`
	AltText   NullString `db:"alt_text"   json:"altText"`
	SortOrder int        `db:"sort_order" json:"sortOrder"`
}

type User struct {
	ID                 string     `db:"id" json:"id"`
	Email              string     `db:"email" json:"email"`
	PasswordHash       string     `db:"password_hash" json:"-"`
	CompanyID          NullString `db:"company_id" json:"companyId"`
	Role               string     `db:"role" json:"role"`
	MustChangePassword bool       `db:"must_change_password" json:"mustChangePassword"`
	CreatedAt          time.Time  `db:"created_at" json:"createdAt"`
}

type QuoteRequest struct {
	ID                string         `db:"id"                  json:"id"`
	RequesterName     string         `db:"requester_name"      json:"requesterName"`
	RequesterCompany  NullString     `db:"requester_company"   json:"requesterCompany"`
	RequesterEmail    string         `db:"requester_email"     json:"requesterEmail"`
	RequesterPhone    NullString     `db:"requester_phone"     json:"requesterPhone"`
	Service           string         `db:"service"             json:"service"`
	Description       NullString     `db:"description"         json:"description"`
	Location          NullString     `db:"location"            json:"location"`
	TargetCompanyID   NullString     `db:"target_company_id"   json:"targetCompanyId"`
	Status            string         `db:"status"              json:"status"`
	ReplyNote         NullString     `db:"reply_note"          json:"replyNote"`
	RepliedAt         *time.Time     `db:"replied_at"          json:"repliedAt"`
	Outcome           NullString     `db:"outcome"             json:"outcome"`
	OutcomeNote       NullString     `db:"outcome_note"        json:"outcomeNote"`
	ClosedAt          *time.Time     `db:"closed_at"           json:"closedAt"`
	FirstResponseAt   *time.Time     `db:"first_response_at"   json:"firstResponseAt"`
	Tags              pq.StringArray `db:"tags"                json:"tags"`
	FollowUpAt        *time.Time     `db:"follow_up_at"        json:"followUpAt"`
	OutcomeAmountCLP  NullInt64      `db:"outcome_amount_clp"  json:"outcomeAmountClp"`
	CreatedAt         time.Time      `db:"created_at"          json:"createdAt"`
}

// Analytics models
type CompanyEvent struct {
	ID        string    `db:"id"         json:"id"`
	CompanyID string    `db:"company_id" json:"-"`
	EventType string    `db:"event_type" json:"eventType"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type DailyMetric struct {
	Date          string `json:"date"`
	ProfileViews  int    `json:"profileViews"`
	ContactClicks int    `json:"contactClicks"`
	RFQs          int    `json:"rfqs"`
}

type AnalyticsTotals struct {
	ProfileViews     int     `json:"profileViews"`
	ContactClicks    int     `json:"contactClicks"`
	QuoteFormOpens   int     `json:"quoteFormOpens"`
	QuoteFormSubmits int     `json:"quoteFormSubmits"`
	RFQsReceived     int     `json:"rfqsReceived"`
	ContactRate      float64 `json:"contactRate"`
}

type AnalyticsResult struct {
	Range  string          `json:"range"`
	Days   int             `json:"days"`
	Totals AnalyticsTotals `json:"totals"`
	Trend  []DailyMetric   `json:"trend"`
}

type CompanyService struct {
	ID          string     `db:"id"          json:"id"`
	CompanyID   string     `db:"company_id"  json:"companyId"`
	Name        string     `db:"name"        json:"name"`
	Category    NullString `db:"category"    json:"category"`
	Description NullString `db:"description" json:"description"`
	Status      string     `db:"status"      json:"status"`
	CreatedAt   time.Time  `db:"created_at"  json:"createdAt"`
}

type ProviderRegistration struct {
	ID          string         `db:"id"           json:"id"`
	CompanyName string         `db:"company_name" json:"companyName"`
	Email       string         `db:"email"        json:"email"`
	Phone       NullString     `db:"phone"        json:"phone"`
	Region      NullString     `db:"region"       json:"region"`
	Services    pq.StringArray `db:"services"     json:"services"`
	Description NullString     `db:"description"  json:"description"`
	Status      string         `db:"status"       json:"status"`
	CreatedAt   time.Time      `db:"created_at"   json:"createdAt"`
}

// Static data types (not stored in DB)
type SubSubcategory struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Subcategory struct {
	Slug        string           `json:"slug"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Icon        string           `json:"icon"`
	Children    []SubSubcategory `json:"children,omitempty"`
}

type CategoryGroup struct {
	Slug          string        `json:"slug"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Icon          string        `json:"icon"`
	Subcategories []Subcategory `json:"subcategories"`
}
