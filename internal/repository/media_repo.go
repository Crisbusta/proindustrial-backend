package repository

import (
	"github.com/crisbusta/proindustrial-backend-public/internal/model"
	"github.com/jmoiron/sqlx"
)

type MediaRepo struct {
	db *sqlx.DB
}

func NewMediaRepo(db *sqlx.DB) *MediaRepo {
	return &MediaRepo{db: db}
}

// ── Company logo / cover ──────────────────────────────────────────────────────

func (r *MediaRepo) UpdateLogo(companyID, url string) error {
	_, err := r.db.Exec(`UPDATE companies SET logo_url=$1, updated_at=NOW() WHERE id=$2`, url, companyID)
	return err
}

func (r *MediaRepo) UpdateCover(companyID, url string) error {
	_, err := r.db.Exec(`UPDATE companies SET cover_url=$1, updated_at=NOW() WHERE id=$2`, url, companyID)
	return err
}

// ── Service images ────────────────────────────────────────────────────────────

func (r *MediaRepo) ListServiceImages(serviceID, companyID string) ([]model.ServiceImage, error) {
	// companyID ownership check via JOIN
	imgs := []model.ServiceImage{}
	err := r.db.Select(&imgs, `
		SELECT si.*
		FROM service_images si
		JOIN company_services cs ON cs.id = si.service_id
		WHERE si.service_id = $1 AND cs.company_id = $2
		ORDER BY si.sort_order ASC, si.created_at ASC
	`, serviceID, companyID)
	return imgs, err
}

func (r *MediaRepo) CountServiceImages(serviceID, companyID string) (int, error) {
	var n int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM service_images si
		JOIN company_services cs ON cs.id = si.service_id
		WHERE si.service_id = $1 AND cs.company_id = $2
	`, serviceID, companyID).Scan(&n)
	return n, err
}

func (r *MediaRepo) AddServiceImage(serviceID, companyID, url string) (*model.ServiceImage, error) {
	// verify ownership first
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM company_services WHERE id=$1 AND company_id=$2)`, serviceID, companyID).Scan(&exists)
	if err != nil || !exists {
		return nil, ErrNotFound
	}

	var img model.ServiceImage
	err = r.db.QueryRowx(`
		INSERT INTO service_images (service_id, url, sort_order)
		VALUES ($1, $2, (SELECT COALESCE(MAX(sort_order)+1, 0) FROM service_images WHERE service_id=$1))
		RETURNING *
	`, serviceID, url).StructScan(&img)
	return &img, err
}

func (r *MediaRepo) DeleteServiceImage(imageID, companyID string) error {
	res, err := r.db.Exec(`
		DELETE FROM service_images si
		USING company_services cs
		WHERE si.id = $1 AND si.service_id = cs.id AND cs.company_id = $2
	`, imageID, companyID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

type ImageOrder struct {
	ID        string `json:"id"`
	SortOrder int    `json:"sortOrder"`
}

func (r *MediaRepo) ReorderServiceImages(serviceID, companyID string, orders []ImageOrder) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, o := range orders {
		_, err := tx.Exec(`
			UPDATE service_images si
			SET sort_order = $1
			FROM company_services cs
			WHERE si.id = $2 AND si.service_id = cs.id AND cs.company_id = $3 AND si.service_id = $4
		`, o.SortOrder, o.ID, companyID, serviceID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ── Certifications ────────────────────────────────────────────────────────────

func (r *MediaRepo) ListCertifications(companyID string) ([]model.CompanyCertification, error) {
	certs := []model.CompanyCertification{}
	err := r.db.Select(&certs, `
		SELECT * FROM company_certifications WHERE company_id=$1 ORDER BY created_at DESC
	`, companyID)
	return certs, err
}

type CertificationInput struct {
	Name        string
	Issuer      string
	DocumentURL string
	IssuedAt    string
	ExpiresAt   string
}

func (r *MediaRepo) CreateCertification(companyID string, in CertificationInput) (*model.CompanyCertification, error) {
	var cert model.CompanyCertification
	err := r.db.QueryRowx(`
		INSERT INTO company_certifications (company_id, name, issuer, document_url, issued_at, expires_at)
		VALUES ($1, $2, NULLIF($3,''), NULLIF($4,''), NULLIF($5,'')::DATE, NULLIF($6,'')::DATE)
		RETURNING *
	`, companyID, in.Name, in.Issuer, in.DocumentURL, in.IssuedAt, in.ExpiresAt).StructScan(&cert)
	return &cert, err
}

func (r *MediaRepo) UpdateCertification(id, companyID string, in CertificationInput) (*model.CompanyCertification, error) {
	var cert model.CompanyCertification
	err := r.db.QueryRowx(`
		UPDATE company_certifications
		SET name=$1, issuer=NULLIF($2,''), document_url=NULLIF($3,''),
		    issued_at=NULLIF($4,'')::DATE, expires_at=NULLIF($5,'')::DATE
		WHERE id=$6 AND company_id=$7
		RETURNING *
	`, in.Name, in.Issuer, in.DocumentURL, in.IssuedAt, in.ExpiresAt, id, companyID).StructScan(&cert)
	if err != nil {
		return nil, ErrNotFound
	}
	return &cert, nil
}

func (r *MediaRepo) DeleteCertification(id, companyID string) error {
	res, err := r.db.Exec(`DELETE FROM company_certifications WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ── Projects ──────────────────────────────────────────────────────────────────

func (r *MediaRepo) ListProjects(companyID string) ([]model.CompanyProject, error) {
	projects := []model.CompanyProject{}
	err := r.db.Select(&projects, `
		SELECT * FROM company_projects WHERE company_id=$1 ORDER BY sort_order ASC, created_at DESC
	`, companyID)
	if err != nil {
		return nil, err
	}
	// attach images
	for i := range projects {
		imgs := []model.ProjectImage{}
		r.db.Select(&imgs, `SELECT * FROM company_project_images WHERE project_id=$1 ORDER BY sort_order ASC`, projects[i].ID)
		projects[i].Images = imgs
	}
	return projects, nil
}

type ProjectInput struct {
	Title       string
	Description string
	ClientName  string
	Year        *int
}

func (r *MediaRepo) CreateProject(companyID string, in ProjectInput) (*model.CompanyProject, error) {
	var p model.CompanyProject
	err := r.db.QueryRowx(`
		INSERT INTO company_projects (company_id, title, description, client_name, year,
		  sort_order)
		VALUES ($1, $2, NULLIF($3,''), NULLIF($4,''),
		  $5,
		  (SELECT COALESCE(MAX(sort_order)+1,0) FROM company_projects WHERE company_id=$1))
		RETURNING *
	`, companyID, in.Title, in.Description, in.ClientName, in.Year).StructScan(&p)
	p.Images = []model.ProjectImage{}
	return &p, err
}

func (r *MediaRepo) UpdateProject(id, companyID string, in ProjectInput) (*model.CompanyProject, error) {
	var p model.CompanyProject
	err := r.db.QueryRowx(`
		UPDATE company_projects
		SET title=$1, description=NULLIF($2,''), client_name=NULLIF($3,''), year=$4
		WHERE id=$5 AND company_id=$6
		RETURNING *
	`, in.Title, in.Description, in.ClientName, in.Year, id, companyID).StructScan(&p)
	if err != nil {
		return nil, ErrNotFound
	}
	imgs := []model.ProjectImage{}
	r.db.Select(&imgs, `SELECT * FROM company_project_images WHERE project_id=$1 ORDER BY sort_order ASC`, p.ID)
	p.Images = imgs
	return &p, nil
}

func (r *MediaRepo) DeleteProject(id, companyID string) error {
	res, err := r.db.Exec(`DELETE FROM company_projects WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MediaRepo) AddProjectImage(projectID, companyID, url string) (*model.ProjectImage, error) {
	var exists bool
	r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM company_projects WHERE id=$1 AND company_id=$2)`, projectID, companyID).Scan(&exists)
	if !exists {
		return nil, ErrNotFound
	}
	var img model.ProjectImage
	err := r.db.QueryRowx(`
		INSERT INTO company_project_images (project_id, url, sort_order)
		VALUES ($1, $2, (SELECT COALESCE(MAX(sort_order)+1,0) FROM company_project_images WHERE project_id=$1))
		RETURNING *
	`, projectID, url).StructScan(&img)
	return &img, err
}

func (r *MediaRepo) DeleteProjectImage(imageID, companyID string) error {
	res, err := r.db.Exec(`
		DELETE FROM company_project_images cpi
		USING company_projects cp
		WHERE cpi.id=$1 AND cpi.project_id=cp.id AND cp.company_id=$2
	`, imageID, companyID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ── Service regions ───────────────────────────────────────────────────────────

func (r *MediaRepo) GetServiceRegions(companyID string) ([]string, error) {
	regions := []string{}
	err := r.db.Select(&regions, `
		SELECT region FROM company_service_regions WHERE company_id=$1 ORDER BY region ASC
	`, companyID)
	return regions, err
}

func (r *MediaRepo) SetServiceRegions(companyID string, regions []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM company_service_regions WHERE company_id=$1`, companyID); err != nil {
		return err
	}
	for _, region := range regions {
		if region == "" {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO company_service_regions (company_id, region) VALUES ($1, $2) ON CONFLICT DO NOTHING`, companyID, region); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ── Public: media for company profile page ────────────────────────────────────

func (r *MediaRepo) GetPublicServiceImages(serviceID string) ([]model.ServiceImage, error) {
	imgs := []model.ServiceImage{}
	err := r.db.Select(&imgs, `SELECT * FROM service_images WHERE service_id=$1 ORDER BY sort_order ASC`, serviceID)
	return imgs, err
}

func (r *MediaRepo) GetPublicCertifications(companyID string) ([]model.CompanyCertification, error) {
	certs := []model.CompanyCertification{}
	err := r.db.Select(&certs, `SELECT * FROM company_certifications WHERE company_id=$1 ORDER BY created_at DESC`, companyID)
	return certs, err
}

func (r *MediaRepo) GetPublicProjects(companyID string) ([]model.CompanyProject, error) {
	return r.ListProjects(companyID)
}
