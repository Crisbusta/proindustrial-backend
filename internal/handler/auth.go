package handler

import (
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authRepo.GetByEmail(body.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":       user.ID,
		"companyId": user.CompanyID,
		"exp":       time.Now().Add(72 * time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	company, _ := h.companyRepo.GetByID(user.CompanyID)
	c.JSON(http.StatusOK, gin.H{
		"token":   tokenStr,
		"userId":  user.ID,
		"company": company,
	})
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
		"userId":  user.ID,
		"email":   user.Email,
		"company": company,
	})
}
