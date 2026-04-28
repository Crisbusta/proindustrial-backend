package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/crisbusta/proindustrial-backend-public/internal/middleware"
	"github.com/crisbusta/proindustrial-backend-public/internal/repository"
	"github.com/crisbusta/proindustrial-backend-public/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxImageSize = 5 << 20  // 5 MB
	maxDocSize   = 10 << 20 // 10 MB
	maxImages    = 8
)

var allowedImageMimes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

var allowedDocMimes = map[string]string{
	"application/pdf": ".pdf",
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
}

type MediaHandler struct {
	repo    *repository.MediaRepo
	company *repository.CompanyRepo
	storage storage.Provider
}

func NewMediaHandler(repo *repository.MediaRepo, company *repository.CompanyRepo, store storage.Provider) *MediaHandler {
	return &MediaHandler{repo: repo, company: company, storage: store}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (h *MediaHandler) uploadFile(c *gin.Context, companyID string, subfolder string, allowed map[string]string, maxBytes int64) (url string, err error) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		return "", fmt.Errorf("archivo requerido")
	}
	defer file.Close()

	if header.Size > maxBytes {
		return "", fmt.Errorf("el archivo supera el límite de %dMB", maxBytes>>20)
	}

	contentType := header.Header.Get("Content-Type")
	// strip charset suffix if present
	contentType = strings.Split(contentType, ";")[0]
	ext, ok := allowed[contentType]
	if !ok {
		supported := make([]string, 0, len(allowed))
		for mime := range allowed {
			supported = append(supported, mime)
		}
		return "", fmt.Errorf("tipo de archivo no permitido (%s). Permitidos: %s", contentType, strings.Join(supported, ", "))
	}

	key := fmt.Sprintf("companies/%s/%s/%s%s", companyID, subfolder, uuid.New().String(), ext)
	url, err = h.storage.Upload(context.Background(), key, file, header.Size, contentType)
	if err != nil {
		slog.Error("storage upload failed", "err", err, "key", key)
		return "", fmt.Errorf("error al guardar el archivo")
	}
	return url, nil
}

func isNotFound(err error) bool {
	return errors.Is(err, repository.ErrNotFound)
}

// ── Profile logo / cover ─────────────────────────────────────────────────────

func (h *MediaHandler) UploadLogo(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	url, err := h.uploadFile(c, companyID, "logo", allowedImageMimes, maxImageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.UpdateLogo(companyID, url); err != nil {
		slog.Error("UpdateLogo db error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *MediaHandler) UploadCover(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	url, err := h.uploadFile(c, companyID, "cover", allowedImageMimes, maxImageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.UpdateCover(companyID, url); err != nil {
		slog.Error("UpdateCover db error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

// ── Service images ────────────────────────────────────────────────────────────

func (h *MediaHandler) ListServiceImages(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	serviceID := c.Param("id")
	imgs, err := h.repo.ListServiceImages(serviceID, companyID)
	if err != nil {
		slog.Error("ListServiceImages error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": imgs})
}

func (h *MediaHandler) AddServiceImage(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	serviceID := c.Param("id")

	count, err := h.repo.CountServiceImages(serviceID, companyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "servicio no encontrado"})
		return
	}
	if count >= maxImages {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("máximo %d imágenes por servicio", maxImages)})
		return
	}

	subfolder := filepath.Join("services", serviceID)
	url, err := h.uploadFile(c, companyID, subfolder, allowedImageMimes, maxImageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	img, err := h.repo.AddServiceImage(serviceID, companyID, url)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "servicio no encontrado"})
			return
		}
		slog.Error("AddServiceImage db error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": img})
}

func (h *MediaHandler) DeleteServiceImage(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	imageID := c.Param("imgId")
	if err := h.repo.DeleteServiceImage(imageID, companyID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "imagen no encontrada"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *MediaHandler) ReorderServiceImages(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	serviceID := c.Param("id")
	var body struct {
		Orders []repository.ImageOrder `json:"orders" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "formato inválido"})
		return
	}
	if err := h.repo.ReorderServiceImages(serviceID, companyID, body.Orders); err != nil {
		slog.Error("ReorderServiceImages error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── Certifications ────────────────────────────────────────────────────────────

func (h *MediaHandler) ListCertifications(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	certs, err := h.repo.ListCertifications(companyID)
	if err != nil {
		slog.Error("ListCertifications error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": certs})
}

func (h *MediaHandler) CreateCertification(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	var body struct {
		Name      string `json:"name" binding:"required"`
		Issuer    string `json:"issuer"`
		IssuedAt  string `json:"issuedAt"`
		ExpiresAt string `json:"expiresAt"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nombre de certificación requerido"})
		return
	}
	cert, err := h.repo.CreateCertification(companyID, repository.CertificationInput{
		Name:      body.Name,
		Issuer:    body.Issuer,
		IssuedAt:  body.IssuedAt,
		ExpiresAt: body.ExpiresAt,
	})
	if err != nil {
		slog.Error("CreateCertification error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": cert})
}

func (h *MediaHandler) UpdateCertification(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	id := c.Param("id")
	var body struct {
		Name      string `json:"name" binding:"required"`
		Issuer    string `json:"issuer"`
		IssuedAt  string `json:"issuedAt"`
		ExpiresAt string `json:"expiresAt"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nombre requerido"})
		return
	}
	cert, err := h.repo.UpdateCertification(id, companyID, repository.CertificationInput{
		Name:      body.Name,
		Issuer:    body.Issuer,
		IssuedAt:  body.IssuedAt,
		ExpiresAt: body.ExpiresAt,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "certificación no encontrada"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cert})
}

func (h *MediaHandler) DeleteCertification(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	if err := h.repo.DeleteCertification(c.Param("id"), companyID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "certificación no encontrada"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *MediaHandler) UploadCertificationDoc(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	id := c.Param("id")

	url, err := h.uploadFile(c, companyID, filepath.Join("certs", id), allowedDocMimes, maxDocSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cert, err := h.repo.UpdateCertification(id, companyID, repository.CertificationInput{DocumentURL: url})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "certificación no encontrada"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cert})
}

// ── Projects ──────────────────────────────────────────────────────────────────

func (h *MediaHandler) ListProjects(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	projects, err := h.repo.ListProjects(companyID)
	if err != nil {
		slog.Error("ListProjects error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": projects})
}

func (h *MediaHandler) CreateProject(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	var body struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		ClientName  string `json:"clientName"`
		Year        *int   `json:"year"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "título del proyecto requerido"})
		return
	}
	p, err := h.repo.CreateProject(companyID, repository.ProjectInput{
		Title:       body.Title,
		Description: body.Description,
		ClientName:  body.ClientName,
		Year:        body.Year,
	})
	if err != nil {
		slog.Error("CreateProject error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": p})
}

func (h *MediaHandler) UpdateProject(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	var body struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		ClientName  string `json:"clientName"`
		Year        *int   `json:"year"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "título requerido"})
		return
	}
	p, err := h.repo.UpdateProject(c.Param("id"), companyID, repository.ProjectInput{
		Title:       body.Title,
		Description: body.Description,
		ClientName:  body.ClientName,
		Year:        body.Year,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "proyecto no encontrado"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

func (h *MediaHandler) DeleteProject(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	if err := h.repo.DeleteProject(c.Param("id"), companyID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "proyecto no encontrado"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *MediaHandler) AddProjectImage(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	projectID := c.Param("id")
	subfolder := filepath.Join("projects", projectID)
	url, err := h.uploadFile(c, companyID, subfolder, allowedImageMimes, maxImageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	img, err := h.repo.AddProjectImage(projectID, companyID, url)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "proyecto no encontrado"})
			return
		}
		slog.Error("AddProjectImage error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": img})
}

func (h *MediaHandler) DeleteProjectImage(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	if err := h.repo.DeleteProjectImage(c.Param("imgId"), companyID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "imagen no encontrada"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── Service regions ───────────────────────────────────────────────────────────

func (h *MediaHandler) GetServiceRegions(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	regions, err := h.repo.GetServiceRegions(companyID)
	if err != nil {
		slog.Error("GetServiceRegions error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"regions": regions})
}

func (h *MediaHandler) UpdateServiceRegions(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	var body struct {
		Regions []string `json:"regions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "regions requerido"})
		return
	}
	if err := h.repo.SetServiceRegions(companyID, body.Regions); err != nil {
		slog.Error("UpdateServiceRegions error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"regions": body.Regions})
}

// ── Public media (no auth) ────────────────────────────────────────────────────

func (h *MediaHandler) GetPublicCertifications(c *gin.Context) {
	slug := c.Param("slug")
	company, err := h.company.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "empresa no encontrada"})
		return
	}
	certs, err := h.repo.GetPublicCertifications(company.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": certs})
}

func (h *MediaHandler) GetPublicProjects(c *gin.Context) {
	slug := c.Param("slug")
	company, err := h.company.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "empresa no encontrada"})
		return
	}
	projects, err := h.repo.GetPublicProjects(company.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": projects})
}
