package handler

import (
	"log/slog"
	"net/http"

	"github.com/crisbusta/proindustrial-backend-public/internal/middleware"
	"github.com/crisbusta/proindustrial-backend-public/internal/repository"
	"github.com/gin-gonic/gin"
)

type PanelHandler struct {
	serviceRepo *repository.ServiceRepo
	quoteRepo   *repository.QuoteRepo
	companyRepo *repository.CompanyRepo
}

func NewPanelHandler(serviceRepo *repository.ServiceRepo, quoteRepo *repository.QuoteRepo, companyRepo *repository.CompanyRepo) *PanelHandler {
	return &PanelHandler{serviceRepo: serviceRepo, quoteRepo: quoteRepo, companyRepo: companyRepo}
}

// Dashboard stats
func (h *PanelHandler) DashboardStats(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)

	stats, err := h.quoteRepo.Stats(companyID)
	if err != nil {
		slog.Error("DashboardStats quoteRepo error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}

	totalServices, err := h.serviceRepo.Count(companyID)
	if err != nil {
		slog.Error("DashboardStats serviceRepo error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	stats.TotalServices = totalServices

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// Profile
func (h *PanelHandler) GetProfile(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	company, err := h.companyRepo.GetByID(companyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": company})
}

func (h *PanelHandler) UpdateProfile(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)

	var body struct {
		Name        *string  `json:"name"`
		Tagline     *string  `json:"tagline"`
		Description *string  `json:"description"`
		Location    *string  `json:"location"`
		Region      *string  `json:"region"`
		Phone       *string  `json:"phone"`
		Email       *string  `json:"email"`
		Website     *string  `json:"website"`
		YearsActive *int     `json:"yearsActive"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos de perfil inválidos"})
		return
	}

	fields := map[string]interface{}{}
	if body.Name != nil        { fields["name"] = *body.Name }
	if body.Tagline != nil     { fields["tagline"] = *body.Tagline }
	if body.Description != nil { fields["description"] = *body.Description }
	if body.Location != nil    { fields["location"] = *body.Location }
	if body.Region != nil      { fields["region"] = *body.Region }
	if body.Phone != nil       { fields["phone"] = *body.Phone }
	if body.Email != nil       { fields["email"] = *body.Email }
	if body.Website != nil     { fields["website"] = *body.Website }
	if body.YearsActive != nil { fields["years_active"] = *body.YearsActive }

	if len(fields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no hay campos para actualizar"})
		return
	}

	if err := h.companyRepo.Update(companyID, fields); err != nil {
		slog.Error("UpdateProfile error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}

	company, _ := h.companyRepo.GetByID(companyID)
	c.JSON(http.StatusOK, gin.H{"data": company})
}

// Services
func (h *PanelHandler) ListServices(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	services, err := h.serviceRepo.List(companyID)
	if err != nil {
		slog.Error("ListServices error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": services})
}

func (h *PanelHandler) CreateService(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)

	var body struct {
		Name        string `json:"name" binding:"required"`
		Category    string `json:"category"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nombre del servicio es requerido"})
		return
	}

	s, err := h.serviceRepo.Create(companyID, repository.CreateServiceInput{
		Name:        body.Name,
		Category:    body.Category,
		Description: body.Description,
	})
	if err != nil {
		slog.Error("CreateService error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": s})
}

func (h *PanelHandler) UpdateService(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	id := c.Param("id")

	var body struct {
		Name        *string `json:"name"`
		Category    *string `json:"category"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos de servicio inválidos"})
		return
	}

	fields := map[string]interface{}{}
	if body.Name != nil        { fields["name"] = *body.Name }
	if body.Category != nil    { fields["category"] = *body.Category }
	if body.Description != nil { fields["description"] = *body.Description }
	if body.Status != nil      { fields["status"] = *body.Status }

	s, err := h.serviceRepo.Update(id, companyID, fields)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": s})
}

func (h *PanelHandler) DeleteService(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	id := c.Param("id")

	if err := h.serviceRepo.Delete(id, companyID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
