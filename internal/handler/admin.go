package handler

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/crisbusta/proindustrial-backend-public/internal/model"
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
	appEnv          string
}

func NewAdminHandler(repo *repository.AdminRepo, mailer *notify.Mailer, initialPassword, appEnv string) *AdminHandler {
	return &AdminHandler{repo: repo, mailer: mailer, initialPassword: initialPassword, appEnv: appEnv}
}

func (h *AdminHandler) generateInitialPassword() string {
	// INITIAL_PASSWORD es una comodidad de desarrollo. En producción daría
	// la misma contraseña a todas las empresas aprobadas, y sus correos son
	// públicos: cualquier proveedor ya aprobado podría entrar a la cuenta
	// de otro antes de su primer ingreso.
	if h.initialPassword != "" && h.appEnv != "production" {
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
		slog.Error("ListRegistrations error", "err", err)
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
		slog.Error("ApproveRegistration bcrypt error", "err", err)
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
			slog.Error("ApproveRegistration error", "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		}
		return
	}

	status, note := h.deliverCredentials(result.Registration.ID, result.User.Email, result.Company.Name, result.InitialPassword)
	result.EmailStatus = status
	result.EmailNote = note

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// emailWaitTimeout: cuánto espera el request por el envío del correo antes
// de responder igual. El commit ya ocurrió, así que colgarse aquí dejaba al
// admin sin la contraseña inicial de una empresa que sí se creó.
const emailWaitTimeout = 12 * time.Second

// deliverBounded ejecuta un envío de correo sin bloquear el request más de
// emailWaitTimeout, y persiste el resultado en el registro para que el admin
// pueda consultarlo después aunque cierre la pantalla. Si el envío se pasa
// del plazo, la goroutine sigue y termina de guardar el estado por su cuenta.
func (h *AdminHandler) deliverBounded(registrationID, slowNote string, send func() notify.DeliveryResult) (string, string) {
	if h.mailer == nil {
		return "logged", "No hay proveedor de correo configurado."
	}

	done := make(chan notify.DeliveryResult, 1)
	go func() {
		delivery := send()
		if err := h.repo.SetEmailStatus(registrationID, delivery.Status, delivery.Note); err != nil {
			slog.Error("SetEmailStatus error", "registrationId", registrationID, "err", err)
		}
		done <- delivery
	}()

	select {
	case delivery := <-done:
		return delivery.Status, delivery.Note
	case <-time.After(emailWaitTimeout):
		return "pending", slowNote
	}
}

func (h *AdminHandler) deliverCredentials(registrationID, email, companyName, password string) (string, string) {
	return h.deliverBounded(
		registrationID,
		"El envío está tardando. Vuelve a cargar la lista en unos minutos para ver si se completó; mientras tanto, entrega la contraseña inicial por otro medio.",
		func() notify.DeliveryResult {
			return h.mailer.SendApprovalEmail(email, companyName, password)
		},
	)
}

// ResendCredentials genera una contraseña inicial nueva para un registro ya
// aprobado y la reenvía. Es la salida cuando el correo no llegó o cuando la
// contraseña se perdió al cerrar la pantalla.
func (h *AdminHandler) ResendCredentials(c *gin.Context) {
	initialPassword := h.generateInitialPassword()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(initialPassword), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("ResendCredentials bcrypt error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}

	result, err := h.repo.ResetInitialPassword(c.Param("id"), string(passwordHash))
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRegistrationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "registro no encontrado"})
		case errors.Is(err, repository.ErrApprovedCompanyNotFound):
			c.JSON(http.StatusConflict, gin.H{"error": "el registro no tiene una empresa aprobada asociada"})
		default:
			slog.Error("ResendCredentials error", "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		}
		return
	}

	result.InitialPassword = initialPassword
	status, note := h.deliverCredentials(result.Registration.ID, result.User.Email, result.Company.Name, initialPassword)
	result.EmailStatus = status
	result.EmailNote = note

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *AdminHandler) RejectRegistration(c *gin.Context) {
	// El cuerpo es opcional: un rechazo sin motivo sigue siendo válido.
	// Notify ausente se interpreta como true, que es la intención por defecto.
	var body struct {
		Reason string `json:"reason"`
		Notify *bool  `json:"notify"`
	}
	_ = c.ShouldBindJSON(&body)
	notifyProvider := body.Notify == nil || *body.Notify

	reason := strings.TrimSpace(body.Reason)
	if len([]rune(reason)) > model.MaxRejectionReason {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("el motivo no puede superar los %d caracteres", model.MaxRejectionReason)})
		return
	}

	reg, err := h.repo.RejectRegistration(c.Param("id"), reason)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRegistrationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "registro no encontrado"})
		case errors.Is(err, repository.ErrRegistrationAlreadyDone):
			c.JSON(http.StatusConflict, gin.H{"error": "el registro ya fue procesado"})
		default:
			slog.Error("RejectRegistration error", "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		}
		return
	}

	if notifyProvider {
		status, note := h.deliverBounded(
			reg.ID,
			"El envío del aviso está tardando. Vuelve a cargar la lista en unos minutos para confirmar si salió.",
			func() notify.DeliveryResult {
				return h.mailer.SendRejectionEmail(reg.Email, reg.CompanyName, reason)
			},
		)
		// Se reflejan en la respuesta para que el admin vea de inmediato si el
		// aviso salió; la versión persistida llega en la siguiente recarga.
		reg.EmailStatus = model.NullString{NullString: sql.NullString{String: status, Valid: true}}
		reg.EmailNote = model.NullString{NullString: sql.NullString{String: note, Valid: note != ""}}
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
			slog.Error("DeleteApprovedCompany error", "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
