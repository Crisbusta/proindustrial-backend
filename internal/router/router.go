package router

import (
	"github.com/crisbusta/proindustrial-backend-public/internal/handler"
	"github.com/crisbusta/proindustrial-backend-public/internal/middleware"
	"github.com/gin-gonic/gin"
)

type Deps struct {
	Company      *handler.CompanyHandler
	Auth         *handler.AuthHandler
	Quote        *handler.QuoteHandler
	Registration *handler.RegistrationHandler
	Panel        *handler.PanelHandler
	JWTSecret    string
	CORSOrigin   string
}

func Setup(deps Deps) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS(deps.CORSOrigin))

	api := r.Group("/api")

	// Public
	api.GET("/category-groups", handler.GetCategoryGroups)
	api.GET("/regions", handler.GetRegions)
	api.GET("/companies", deps.Company.List)
	api.GET("/companies/:slug", deps.Company.GetBySlug)
	api.POST("/quotes", deps.Quote.Create)
	api.POST("/registrations", deps.Registration.Create)

	// Auth
	api.POST("/auth/login", deps.Auth.Login)
	api.GET("/auth/me", middleware.Auth(deps.JWTSecret), deps.Auth.Me)

	// Panel (protected)
	panel := api.Group("/panel", middleware.Auth(deps.JWTSecret))
	panel.GET("/dashboard/stats", deps.Panel.DashboardStats)
	panel.GET("/quotes", deps.Quote.List)
	panel.PATCH("/quotes/:id", deps.Quote.UpdateStatus)
	panel.GET("/services", deps.Panel.ListServices)
	panel.POST("/services", deps.Panel.CreateService)
	panel.PATCH("/services/:id", deps.Panel.UpdateService)
	panel.DELETE("/services/:id", deps.Panel.DeleteService)
	panel.GET("/profile", deps.Panel.GetProfile)
	panel.PUT("/profile", deps.Panel.UpdateProfile)

	return r
}
