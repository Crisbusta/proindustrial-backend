package handler

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/crisbusta/proindustrial-backend-public/internal/middleware"
	"github.com/crisbusta/proindustrial-backend-public/internal/repository"
	"github.com/gin-gonic/gin"
)

var validEventTypes = map[string]bool{
	"profile_view":          true,
	"service_view":          true,
	"contact_click_phone":   true,
	"contact_click_whatsapp": true,
	"contact_click_email":   true,
	"quote_form_open":       true,
	"quote_form_submit":     true,
}

const ipSalt = "proindustrial-2026"

type AnalyticsHandler struct {
	events    *repository.EventRepo
	companies *repository.CompanyRepo
}

func NewAnalyticsHandler(events *repository.EventRepo, companies *repository.CompanyRepo) *AnalyticsHandler {
	return &AnalyticsHandler{events: events, companies: companies}
}

// POST /api/events — public, no auth
func (h *AnalyticsHandler) TrackEvent(c *gin.Context) {
	var body struct {
		CompanyID string `json:"companyId" binding:"required"`
		EventType string `json:"eventType" binding:"required"`
		VisitorID string `json:"visitorId"`
		Referrer  string `json:"referrer"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "companyId y eventType requeridos"})
		return
	}
	if !validEventTypes[body.EventType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "eventType inválido"})
		return
	}

	ip := c.ClientIP()
	ipHash := fmt.Sprintf("%x", sha256.Sum256([]byte(ip+ipSalt)))

	if err := h.events.Insert(body.CompanyID, body.EventType, body.VisitorID, body.Referrer, ipHash); err != nil {
		slog.Error("TrackEvent insert error", "err", err)
	}
	c.Status(http.StatusNoContent)
}

// GET /panel/analytics?range=7d|30d|90d
func (h *AnalyticsHandler) GetAnalytics(c *gin.Context) {
	companyID := c.GetString(middleware.CompanyIDKey)
	days := 30
	switch strings.TrimSpace(c.Query("range")) {
	case "7d":
		days = 7
	case "90d":
		days = 90
	}

	result, err := h.events.GetAnalytics(companyID, days)
	if err != nil {
		slog.Error("GetAnalytics error", "err", err, "company_id", companyID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
