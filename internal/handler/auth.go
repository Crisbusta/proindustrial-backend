package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/crisbusta/proindustrial-backend-public/internal/middleware"
	"github.com/crisbusta/proindustrial-backend-public/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	authRepo    *repository.AuthRepo
	companyRepo *repository.CompanyRepo
	jwtSecret   string
}

func NewAuthHandler(authRepo *repository.AuthRepo, companyRepo *repository.CompanyRepo, jwtSecret string) *AuthHandler {
	return &AuthHandler{authRepo: authRepo, companyRepo: companyRepo, jwtSecret: jwtSecret}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "correo y contraseña son requeridos"})
		return
	}

	email := body.Email
	password := body.Password
	resp, err := h.login(email, password, "provider")
	if err != nil {
		c.JSON(err.status, gin.H{"error": err.message})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var body struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contraseña actual y nueva son requeridas (mínimo 8 caracteres)"})
		return
	}

	user, err := h.authRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current password is incorrect"})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("ChangePassword bcrypt error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}

	if err := h.authRepo.ChangePassword(user.ID, string(passwordHash)); err != nil {
		slog.Error("ChangePassword repo error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}

	// Issue a fresh token with mustChangePassword = false
	newClaims := jwt.MapClaims{
		"sub":                user.ID,
		"role":               user.Role,
		"exp":                time.Now().Add(72 * time.Hour).Unix(),
		"mustChangePassword": false,
	}
	if user.CompanyID.Valid {
		newClaims["companyId"] = user.CompanyID.String
	}
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenStr, err := newToken.SignedString([]byte(h.jwtSecret))
	if err != nil {
		slog.Error("ChangePassword token sign error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "token": newTokenStr})
}

func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "correo y contraseña son requeridos"})
		return
	}

	resp, err := h.login(body.Email, body.Password, "admin")
	if err != nil {
		c.JSON(err.status, gin.H{"error": err.message})
		return
	}
	c.JSON(http.StatusOK, resp)
}

type loginError struct {
	status  int
	message string
}

func (h *AuthHandler) login(email, password, expectedRole string) (gin.H, *loginError) {
	user, err := h.authRepo.GetByEmail(email)
	if err != nil {
		return nil, &loginError{status: http.StatusUnauthorized, message: "invalid credentials"}
	}

	if user.Role != expectedRole {
		return nil, &loginError{status: http.StatusForbidden, message: "forbidden"}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, &loginError{status: http.StatusUnauthorized, message: "invalid credentials"}
	}

	claims := jwt.MapClaims{
		"sub":                user.ID,
		"role":               user.Role,
		"exp":                time.Now().Add(72 * time.Hour).Unix(),
		"mustChangePassword": user.MustChangePassword,
	}
	if user.CompanyID.Valid {
		claims["companyId"] = user.CompanyID.String
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		slog.Error("login token sign error", "err", err)
		return nil, &loginError{status: http.StatusInternalServerError, message: "error interno del servidor"}
	}

	response := gin.H{
		"token":              tokenStr,
		"userId":             user.ID,
		"email":              user.Email,
		"role":               user.Role,
		"mustChangePassword": user.MustChangePassword,
	}
	if user.CompanyID.Valid {
		company, _ := h.companyRepo.GetByID(user.CompanyID.String)
		response["company"] = company
	}
	return response, nil
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	companyID := c.GetString(middleware.CompanyIDKey)

	user, err := h.authRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	company, _ := h.companyRepo.GetByID(companyID)
	c.JSON(http.StatusOK, gin.H{
		"userId":             user.ID,
		"email":              user.Email,
		"role":               user.Role,
		"mustChangePassword": user.MustChangePassword,
		"company":            company,
	})
}

func (h *AuthHandler) AdminMe(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	user, err := h.authRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userId":             user.ID,
		"email":              user.Email,
		"role":               user.Role,
		"mustChangePassword": user.MustChangePassword,
	})
}
