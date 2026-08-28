package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/crisbusta/proindustrial-backend-public/internal/model"
	"github.com/crisbusta/proindustrial-backend-public/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type RegistrationHandler struct {
	repo *repository.RegistrationRepo
}

func NewRegistrationHandler(repo *repository.RegistrationRepo) *RegistrationHandler {
	return &RegistrationHandler{repo: repo}
}

// controlChars: Postgres no admite \x00 en columnas text, y el resto de
// caracteres de control se guardan pero rompen el render. Se eliminan
// antes de tocar la base para no devolver un 500 opaco.
var controlChars = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)

var multiSpace = regexp.MustCompile(`\s{2,}`)

func sanitizeText(s string) string {
	return strings.TrimSpace(controlChars.ReplaceAllString(s, ""))
}

// fieldLabels traduce el nombre del campo Go al término que ve el usuario.
var fieldLabels = map[string]string{
	"CompanyName": "el nombre de la empresa",
	"Email":       "el correo electrónico",
	"Phone":       "el teléfono",
	"Region":      "la región",
	"Description": "la descripción",
	"Services":    "los servicios",
}

// validationMessage convierte los errores del validador en un mensaje en
// español que nombra el campo concreto. Antes, cualquier fallo de binding
// devolvía siempre "nombre de empresa y correo son requeridos", que era
// engañoso cuando el problema era otro.
func validationMessage(err error) string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return "el formato de la solicitud no es válido"
	}

	msgs := make([]string, 0, len(ve))
	for _, fe := range ve {
		label, ok := fieldLabels[fe.Field()]
		if !ok {
			label = "un campo"
		}
		switch fe.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s es obligatorio", label))
		case "email":
			msgs = append(msgs, "el correo electrónico no tiene un formato válido")
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s debe tener al menos %s caracteres", label, fe.Param()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s no puede superar %s caracteres", label, fe.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s no es válido", label))
		}
	}
	return strings.Join(msgs, "; ")
}

// cleanServices descarta cualquier slug que no exista en el catálogo y
// elimina repetidos. Sin esto, un POST directo podía sembrar categorías
// inventadas que después se publicaban en la ficha de la empresa.
func cleanServices(raw []string) []string {
	valid := make(map[string]bool, len(repository.CategoryGroups))
	for _, g := range repository.CategoryGroups {
		valid[g.Slug] = true
	}

	seen := make(map[string]bool, len(raw))
	out := make([]string, 0, len(raw))
	for _, s := range raw {
		slug := strings.ToLower(strings.TrimSpace(s))
		if valid[slug] && !seen[slug] {
			seen[slug] = true
			out = append(out, slug)
		}
	}
	return out
}

// hasDomainDot exige un punto en el dominio. El tag "email" del validador
// sigue la RFC y acepta direcciones como "a@localhost", a las que nunca
// llegaría el correo con la contraseña inicial.
func hasDomainDot(email string) bool {
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return false
	}
	return strings.Contains(email[at+1:], ".")
}

func isKnownRegion(region string) bool {
	for _, r := range repository.Regions {
		if r != "Todas las regiones" && r == region {
			return true
		}
	}
	return false
}

func (h *RegistrationHandler) Create(c *gin.Context) {
	var body struct {
		CompanyName string   `json:"companyName" binding:"required"`
		Email       string   `json:"email"       binding:"required,email"`
		Phone       string   `json:"phone"`
		Region      string   `json:"region"`
		Services    []string `json:"services"`
		Description string   `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationMessage(err)})
		return
	}

	// Normalizar antes de validar longitudes: el correo se guarda siempre
	// en minúsculas porque el login lo compara así, y un correo con
	// mayúsculas dejaba al proveedor aprobado sin poder entrar nunca.
	companyName := multiSpace.ReplaceAllString(sanitizeText(body.CompanyName), " ")
	email := strings.ToLower(sanitizeText(body.Email))
	phone := sanitizeText(body.Phone)
	region := sanitizeText(body.Region)
	description := sanitizeText(body.Description)
	services := cleanServices(body.Services)

	if n := len([]rune(companyName)); n < model.MinCompanyName || n > model.MaxCompanyName {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("el nombre de la empresa debe tener entre %d y %d caracteres", model.MinCompanyName, model.MaxCompanyName)})
		return
	}
	if len([]rune(email)) > model.MaxEmail || !hasDomainDot(email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el correo electrónico no tiene un formato válido"})
		return
	}
	if n := len([]rune(phone)); n < model.MinPhone || n > model.MaxPhone {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("el teléfono debe tener entre %d y %d caracteres", model.MinPhone, model.MaxPhone)})
		return
	}
	if !isKnownRegion(region) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "selecciona una región válida"})
		return
	}
	if n := len([]rune(description)); n < model.MinDescription || n > model.MaxDescription {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("la descripción debe tener entre %d y %d caracteres", model.MinDescription, model.MaxDescription)})
		return
	}
	if len(services) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "selecciona al menos una categoría de servicio"})
		return
	}
	if len(services) > model.MaxServices {
		services = services[:model.MaxServices]
	}

	// Un proveedor que reenvía el formulario tras un timeout no debe
	// generar dos fichas idénticas en la cola del admin.
	if dup, err := h.repo.PendingExists(email); err == nil && dup {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Ya recibimos una solicitud con este correo y está en revisión. Te contactaremos pronto."})
		return
	}

	reg, err := h.repo.Create(repository.CreateRegistrationInput{
		CompanyName: companyName,
		Email:       email,
		Phone:       phone,
		Region:      region,
		Services:    services,
		Description: description,
	})
	if err != nil {
		if repository.IsUniqueViolation(err) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Ya recibimos una solicitud con este correo y está en revisión. Te contactaremos pronto."})
			return
		}
		slog.Error("Registration.Create error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": reg})
}
