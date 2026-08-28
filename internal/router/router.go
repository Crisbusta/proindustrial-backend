package router

import (
	"log/slog"
	"net/http"

	"github.com/crisbusta/proindustrial-backend-public/internal/handler"
	"github.com/crisbusta/proindustrial-backend-public/internal/middleware"
	"github.com/crisbusta/proindustrial-backend-public/internal/model"
	"github.com/crisbusta/proindustrial-backend-public/internal/storage"
	"github.com/gin-gonic/gin"
)

type Deps struct {
	Company      *handler.CompanyHandler
	Auth         *handler.AuthHandler
	Quote        *handler.QuoteHandler
	Registration *handler.RegistrationHandler
	Panel        *handler.PanelHandler
	Admin        *handler.AdminHandler
	Health       *handler.HealthHandler
	Media        *handler.MediaHandler
	Analytics    *handler.AnalyticsHandler
	Storage      storage.Provider
	StorageDir   string
	JWTSecret    string
	CORSOrigin   string

	TrustedPlatform string
	TrustedProxies  []string
}

func Setup(deps Deps) *gin.Engine {
	r := gin.New()

	// Por defecto gin acepta el X-Forwarded-For de cualquier origen, lo que
	// permite evadir el rate limit por IP rotando la cabecera. Declarar la
	// plataforma o los proxies reales hace que ClientIP() no sea falsificable.
	switch {
	case deps.TrustedPlatform != "":
		r.TrustedPlatform = deps.TrustedPlatform
	case len(deps.TrustedProxies) > 0:
		if err := r.SetTrustedProxies(deps.TrustedProxies); err != nil {
			slog.Error("SetTrustedProxies inválido, se mantiene el comportamiento por defecto", "err", err)
		}
	default:
		slog.Warn("TRUSTED_PLATFORM / TRUSTED_PROXIES sin configurar: el rate limit por IP se puede evadir falsificando X-Forwarded-For")
	}

	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(deps.CORSOrigin))
	r.Use(middleware.Security())
	r.Use(middleware.BodyLimit(25 << 20)) // 25 MB: techo global, deja pasar las subidas de imágenes

	// Health (no auth, no rate limit)
	r.GET("/healthz", deps.Health.Healthz)
	r.GET("/readyz", deps.Health.Readyz)

	// Static uploads — only used with the local storage driver.
	// When STORAGE_DRIVER=s3, files are served directly from S3/R2/CDN
	// and this route is never hit (PublicURL points elsewhere).
	if deps.StorageDir != "" {
		r.StaticFS("/uploads", http.Dir(deps.StorageDir))
	}

	api := r.Group("/api")

	// Public
	publicWrite := []gin.HandlerFunc{middleware.PublicWriteLimit(), middleware.BodyLimit(model.MaxBodyBytes)}

	// /events no lleva PublicWriteLimit: se dispara en cada vista de página
	// y un límite estrecho devolvería 429 a la navegación normal.
	api.POST("/events", middleware.EventLimit(), middleware.BodyLimit(model.MaxBodyBytes), deps.Analytics.TrackEvent)
	api.GET("/category-groups", handler.GetCategoryGroups)
	api.GET("/regions", handler.GetRegions)
	api.GET("/companies", deps.Company.List)
	api.GET("/companies/:slug", deps.Company.GetBySlug)
	api.GET("/companies/:slug/services", deps.Company.ListServices)
	api.GET("/companies/:slug/certifications", deps.Media.GetPublicCertifications)
	api.GET("/companies/:slug/projects", deps.Media.GetPublicProjects)
	api.POST("/quotes", append(publicWrite, deps.Quote.Create)...)
	api.POST("/registrations", append(publicWrite, deps.Registration.Create)...)

	// Auth (rate-limited)
	api.POST("/auth/login", middleware.RateLimit(), deps.Auth.Login)
	api.GET("/auth/me", middleware.Auth(deps.JWTSecret, "provider"), deps.Auth.Me)
	api.POST("/auth/change-password", middleware.Auth(deps.JWTSecret, "provider"), deps.Auth.ChangePassword)

	// Panel (protected)
	panel := api.Group("/panel", middleware.Auth(deps.JWTSecret, "provider"))
	panel.GET("/dashboard/stats", deps.Panel.DashboardStats)
	panel.GET("/quotes", deps.Quote.List)
	panel.PATCH("/quotes/:id", deps.Quote.UpdateStatus)
	panel.POST("/quotes/:id/reply", deps.Quote.Reply)
	panel.POST("/quotes/:id/close", deps.Quote.Close)
	panel.PATCH("/quotes/:id/tags", deps.Quote.SetTags)
	panel.PATCH("/quotes/:id/follow-up", deps.Quote.SetFollowUp)
	panel.GET("/analytics", deps.Analytics.GetAnalytics)
	panel.GET("/services", deps.Panel.ListServices)
	panel.POST("/services", deps.Panel.CreateService)
	panel.PATCH("/services/:id", deps.Panel.UpdateService)
	panel.DELETE("/services/:id", deps.Panel.DeleteService)
	panel.GET("/profile", deps.Panel.GetProfile)
	panel.PUT("/profile", deps.Panel.UpdateProfile)

	// Media — profile images
	panel.POST("/profile/logo", deps.Media.UploadLogo)
	panel.POST("/profile/cover", deps.Media.UploadCover)
	panel.GET("/profile/regions", deps.Media.GetServiceRegions)
	panel.PUT("/profile/regions", deps.Media.UpdateServiceRegions)

	// Media — service images
	panel.GET("/services/:id/images", deps.Media.ListServiceImages)
	panel.POST("/services/:id/images", deps.Media.AddServiceImage)
	panel.DELETE("/services/:id/images/:imgId", deps.Media.DeleteServiceImage)
	panel.PATCH("/services/:id/images/reorder", deps.Media.ReorderServiceImages)

	// Media — certifications
	panel.GET("/certifications", deps.Media.ListCertifications)
	panel.POST("/certifications", deps.Media.CreateCertification)
	panel.PATCH("/certifications/:id", deps.Media.UpdateCertification)
	panel.DELETE("/certifications/:id", deps.Media.DeleteCertification)
	panel.POST("/certifications/:id/document", deps.Media.UploadCertificationDoc)

	// Media — projects
	panel.GET("/projects", deps.Media.ListProjects)
	panel.POST("/projects", deps.Media.CreateProject)
	panel.PATCH("/projects/:id", deps.Media.UpdateProject)
	panel.DELETE("/projects/:id", deps.Media.DeleteProject)
	panel.POST("/projects/:id/images", deps.Media.AddProjectImage)
	panel.DELETE("/projects/:id/images/:imgId", deps.Media.DeleteProjectImage)

	// Admin auth (rate-limited)
	adminAuth := api.Group("/admin/auth")
	adminAuth.POST("/login", middleware.RateLimit(), deps.Auth.AdminLogin)
	adminAuth.GET("/me", middleware.Auth(deps.JWTSecret, "admin"), deps.Auth.AdminMe)
	adminAuth.POST("/change-password", middleware.Auth(deps.JWTSecret, "admin"), deps.Auth.ChangePassword)

	// Admin (protected)
	admin := api.Group("/admin", middleware.Auth(deps.JWTSecret, "admin"))
	admin.GET("/registrations", deps.Admin.ListRegistrations)
	admin.GET("/registrations/:id", deps.Admin.GetRegistration)
	admin.POST("/registrations/:id/approve", deps.Admin.ApproveRegistration)
	admin.POST("/registrations/:id/reject", deps.Admin.RejectRegistration)
	admin.POST("/registrations/:id/resend-credentials", deps.Admin.ResendCredentials)
	admin.DELETE("/registrations/:id/company", deps.Admin.DeleteApprovedCompany)

	return r
}
