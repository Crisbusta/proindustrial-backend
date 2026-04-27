package handler

import (
	"log/slog"
	"net/http"

	"github.com/crisbusta/proindustrial-backend-public/internal/repository"
	"github.com/gin-gonic/gin"
)

type RegistrationHandler struct {
	repo *repository.RegistrationRepo
}

func NewRegistrationHandler(repo *repository.RegistrationRepo) *RegistrationHandler {
	return &RegistrationHandler{repo: repo}
}

func (h *RegistrationHandler) Create(c *gin.Context) {
	var body struct {
		CompanyName string   `json:"companyName" binding:"required"`
		Email       string   `json:"email" binding:"required,email"`
		Phone       string   `json:"phone"`
		Region      string   `json:"region"`
		Services    []string `json:"services"`
		Description string   `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nombre de empresa y correo son requeridos"})
		return
	}

	reg, err := h.repo.Create(repository.CreateRegistrationInput{
		CompanyName: body.CompanyName,
		Email:       body.Email,
		Phone:       body.Phone,
		Region:      body.Region,
		Services:    body.Services,
		Description: body.Description,
	})
	if err != nil {
		slog.Error("Registration.Create error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": reg})
}
