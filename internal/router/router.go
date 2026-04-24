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
	Admin        *handler.AdminHandler
	JWTSecret    string
	CORSOrigin   string
}

func Setup(deps Deps) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS(deps.CORSOrigin))
	r.Use(middleware.Security())

	api := r.Group("/api")

	// Public
	api.GET("/category-groups", handler.GetCategoryGroups)
	api.GET("/regions", handler.GetRegions)
	api.GET("/companies", deps.Company.List)
	api.GET("/companies/:slug", deps.Company.GetBySlug)
	api.GET("/companies/:slug/services", deps.Company.ListServices)
	api.POST("/quotes", deps.Quote.Create)
	api.POST("/registrations", deps.Registration.Create)

	// Auth
	api.POST("/auth/login", deps.Auth.Login)
	api.GET("/auth/me", middleware.Auth(deps.JWTSecret, "provider"), deps.Auth.Me)
	api.POST("/auth/change-password", middleware.Auth(deps.JWTSecret, "provider"), deps.Auth.ChangePassword)

	// Panel (protected)
	panel := api.Group("/panel", middleware.Auth(deps.JWTSecret, "provider"))
	panel.GET("/dashboard/stats", deps.Panel.DashboardStats)
	panel.GET("/quotes", deps.Quote.List)
	panel.PATCH("/quotes/:id", deps.Quote.UpdateStatus)
	panel.POST("/quotes/:id/reply", deps.Quote.Reply)
	panel.POST("/quotes/:id/close", deps.Quote.Close)
	panel.GET("/services", deps.Panel.ListServices)
	panel.POST("/services", deps.Panel.CreateService)
	panel.PATCH("/services/:id", deps.Panel.UpdateService)
	panel.DELETE("/services/:id", deps.Panel.DeleteService)
	panel.GET("/profile", deps.Panel.GetProfile)
	panel.PUT("/profile", deps.Panel.UpdateProfile)

	// Admin
	adminAuth := api.Group("/admin/auth")
	adminAuth.POST("/login", deps.Auth.AdminLogin)
	adminAuth.GET("/me", middleware.Auth(deps.JWTSecret, "admin"), deps.Auth.AdminMe)
	adminAuth.POST("/change-password", middleware.Auth(deps.JWTSecret, "admin"), deps.Auth.ChangePassword)

	admin := api.Group("/admin", middleware.Auth(deps.JWTSecret, "admin"))
	admin.GET("/registrations", deps.Admin.ListRegistrations)
	admin.GET("/registrations/:id", deps.Admin.GetRegistration)
	admin.POST("/registrations/:id/approve", deps.Admin.ApproveRegistration)
	admin.POST("/registrations/:id/reject", deps.Admin.RejectRegistration)
	admin.DELETE("/registrations/:id/company", deps.Admin.DeleteApprovedCompany)

	return r
}
