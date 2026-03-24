package handler

import (
	"crypto/rand"
	"errors"
	"log"
	"math/big"
	"net/http"

	"github.com/crisbusta/proindustrial-backend-public/internal/notify"
	"github.com/crisbusta/proindustrial-backend-public/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const passwordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type AdminHandler struct {
	repo            *repository.AdminRepo
	mailer          *notify.Mailer
	initialPassword string
}

func NewAdminHandler(repo *repository.AdminRepo, mailer *notify.Mailer, initialPassword string) *AdminHandler {
	return &AdminHandler{repo: repo, mailer: mailer, initialPassword: initialPassword}
}

func (h *AdminHandler) generateInitialPassword() string {
	if h.initialPassword != "" {
		return h.initialPassword
	}
	b := make([]byte, 12)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordChars))))
		if err != nil {
			// fallback: should not happen
			b[i] = passwordChars[i%len(passwordChars)]
			continue
		}
		b[i] = passwordChars[n.Int64()]
	}
	return string(b)
}

func (h *AdminHandler) ListRegistrations(c *gin.Context) {
	status := c.Query("status")
	regs, err := h.repo.ListRegistrations(status)
	if err != nil {
		log.Printf("ListRegistrations error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": regs})
}

func (h *AdminHandler) GetRegistration(c *gin.Context) {
	reg, err := h.repo.GetRegistrationByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "registro no encontrado"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": reg})
}

func (h *AdminHandler) ApproveRegistration(c *gin.Context) {
	initialPassword := h.generateInitialPassword()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(initialPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ApproveRegistration bcrypt error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}

	result, err := h.repo.ApproveRegistration(c.Param("id"), string(passwordHash), initialPassword)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRegistrationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "registro no encontrado"})
		case errors.Is(err, repository.ErrRegistrationAlreadyDone):
			c.JSON(http.StatusConflict, gin.H{"error": "el registro ya fue procesado"})
		case errors.Is(err, repository.ErrRegistrationEmailInUse):
			c.JSON(http.StatusConflict, gin.H{"error": "el correo ya está en uso"})
		default:
			log.Printf("ApproveRegistration error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
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
			c.JSON(http.StatusNotFound, gin.H{"error": "registro no encontrado"})
		case errors.Is(err, repository.ErrRegistrationAlreadyDone):
			c.JSON(http.StatusConflict, gin.H{"error": "el registro ya fue procesado"})
		default:
			log.Printf("RejectRegistration error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
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
			c.JSON(http.StatusNotFound, gin.H{"error": "registro no encontrado"})
		case errors.Is(err, repository.ErrApprovedCompanyNotFound):
			c.JSON(http.StatusConflict, gin.H{"error": "empresa aprobada no encontrada para este registro"})
		default:
			log.Printf("DeleteApprovedCompany error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
