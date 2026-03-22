package handler

import (
	"net/http"

	"github.com/crisbusta/proindustrial-backend-public/internal/repository"
	"github.com/gin-gonic/gin"
)

func GetCategoryGroups(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": repository.CategoryGroups})
}

func GetRegions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": repository.Regions})
}
