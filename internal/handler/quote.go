package handler

import (
	"log"
	"net/http"

	"github.com/crisbusta/proindustrial-backend-public/internal/middleware"
	"github.com/crisbusta/proindustrial-backend-public/internal/repository"
	"github.com/gin-gonic/gin"
)

type QuoteHandler struct {
	repo *repository.QuoteRepo
}

func NewQuoteHandler(repo *repository.QuoteRepo) *QuoteHandler {
	return &QuoteHandler{repo: repo}
}

func (h *QuoteHandler) Create(c *gin.Context) {
	var body struct {
		RequesterName    string `json:"requesterName" binding:"required"`
		RequesterCompany string `json:"requesterCompany"`
		RequesterEmail   string `json:"requesterEmail" binding:"required,email"`
		RequesterPhone   string `json:"requesterPhone"`
		Service          string `json:"service" binding:"required"`
		Description      string `json:"description"`
		Location         string `json:"location"`
		TargetCompanyID  string `json:"targetCompanyId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nombre, correo y servicio son requeridos"})
		return
	}

	q, err := h.repo.Create(repository.CreateQuoteInput{
		RequesterName:    body.RequesterName,
		RequesterCompany: body.RequesterCompany,
		RequesterEmail:   body.RequesterEmail,
		RequesterPhone:   body.RequesterPhone,
		Service:          body.Service,
		Description:      body.Description,
		Location:         body.Location,
		TargetCompanyID:  body.TargetCompanyID,
	})
	if err != nil {
		log.Printf("Quote.Create error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": q})
}

func (h *QuoteHandler) List(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	status := c.Query("status")

	quotes, err := h.repo.ListByCompany(companyID, status)
	if err != nil {
		log.Printf("Quote.List error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": quotes})
}

func (h *QuoteHandler) UpdateStatus(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	id := c.Param("id")

	var body struct {
		Status string `json:"status" binding:"required,oneof=new read responded"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateStatus(id, companyID, body.Status); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quote not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
