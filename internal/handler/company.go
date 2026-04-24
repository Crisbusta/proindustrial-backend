package handler

import (
	"net/http"

	"github.com/crisbusta/proindustrial-backend-public/internal/repository"
	"github.com/gin-gonic/gin"
)

type CompanyHandler struct {
	repo        *repository.CompanyRepo
	serviceRepo *repository.ServiceRepo
}

func NewCompanyHandler(repo *repository.CompanyRepo, serviceRepo *repository.ServiceRepo) *CompanyHandler {
	return &CompanyHandler{repo: repo, serviceRepo: serviceRepo}
}

func (h *CompanyHandler) List(c *gin.Context) {
	category := c.Query("category")
	region := c.Query("region")

	var featured *bool
	if f := c.Query("featured"); f == "true" {
		t := true
		featured = &t
	} else if f == "false" {
		f2 := false
		featured = &f2
	}

	companies, err := h.repo.List(category, region, featured)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": companies})
}

func (h *CompanyHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	company, err := h.repo.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": company})
}

func (h *CompanyHandler) ListServices(c *gin.Context) {
	slug := c.Param("slug")
	company, err := h.repo.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}
	services, err := h.serviceRepo.ListActive(company.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": services})
}
