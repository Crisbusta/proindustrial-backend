package handler

import (
	"errors"
	"net/http"

	"github.com/crisbusta/proindustrial-backend-public/internal/notify"
	"github.com/crisbusta/proindustrial-backend-public/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const approvedRegistrationPassword = "demo123"

type AdminHandler struct {
	repo   *repository.AdminRepo
	mailer *notify.Mailer
}

func NewAdminHandler(repo *repository.AdminRepo, mailer *notify.Mailer) *AdminHandler {
	return &AdminHandler{repo: repo, mailer: mailer}
}

func (h *AdminHandler) ListRegistrations(c *gin.Context) {
	status := c.Query("status")
	regs, err := h.repo.ListRegistrations(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": regs})
}

func (h *AdminHandler) GetRegistration(c *gin.Context) {
	reg, err := h.repo.GetRegistrationByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "registration not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": reg})
}

func (h *AdminHandler) ApproveRegistration(c *gin.Context) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(approvedRegistrationPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create password"})
		return
	}

	result, err := h.repo.ApproveRegistration(c.Param("id"), string(passwordHash), approvedRegistrationPassword)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRegistrationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "registration not found"})
		case errors.Is(err, repository.ErrRegistrationAlreadyDone):
			c.JSON(http.StatusConflict, gin.H{"error": "registration already processed"})
		case errors.Is(err, repository.ErrRegistrationEmailInUse):
			c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if h.mailer != nil {
		delivery := h.mailer.SendApprovalEmail(result.User.Email, result.Company.Name, result.InitialPassword)
		result.EmailStatus = delivery.Status
		result.EmailNote = delivery.Note
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *AdminHandler) RejectRegistration(c *gin.Context) {
	reg, err := h.repo.RejectRegistration(c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRegistrationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "registration not found"})
		case errors.Is(err, repository.ErrRegistrationAlreadyDone):
			c.JSON(http.StatusConflict, gin.H{"error": "registration already processed"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": reg})
}

func (h *AdminHandler) DeleteApprovedCompany(c *gin.Context) {
	err := h.repo.DeleteApprovedCompanyByRegistration(c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRegistrationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "registration not found"})
		case errors.Is(err, repository.ErrApprovedCompanyNotFound):
			c.JSON(http.StatusConflict, gin.H{"error": "approved company not found for this registration"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
