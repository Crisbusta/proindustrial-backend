package main

import (
	"log/slog"

	"github.com/crisbusta/proindustrial-backend-public/internal/config"
	"github.com/crisbusta/proindustrial-backend-public/internal/database"
	"github.com/crisbusta/proindustrial-backend-public/internal/handler"
	"github.com/crisbusta/proindustrial-backend-public/internal/logger"
	"github.com/crisbusta/proindustrial-backend-public/internal/notify"
	"github.com/crisbusta/proindustrial-backend-public/internal/repository"
	"github.com/crisbusta/proindustrial-backend-public/internal/router"
	"github.com/crisbusta/proindustrial-backend-public/internal/storage"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.AppEnv)

	store, err := storage.New(storage.Config{
		Driver:      cfg.StorageDriver,
		BaseURL:     cfg.AppBaseURL + "/uploads",
		Dir:         cfg.StorageDir,
		S3Bucket:    cfg.S3Bucket,
		S3Region:    cfg.S3Region,
		S3Endpoint:  cfg.S3Endpoint,
		S3AccessKey: cfg.S3AccessKey,
		S3SecretKey: cfg.S3SecretKey,
		S3PublicBase: cfg.S3PublicBase,
	})
	if err != nil {
		slog.Error("storage init failed", "err", err)
	}

	mailer := notify.NewMailer(cfg)

	db := database.Connect(cfg.DatabaseURL)
	database.RunMigrations(cfg.DatabaseURL)

	// Repos
	companyRepo := repository.NewCompanyRepo(db)
	authRepo := repository.NewAuthRepo(db)
	quoteRepo := repository.NewQuoteRepo(db)
	serviceRepo := repository.NewServiceRepo(db)
	registrationRepo := repository.NewRegistrationRepo(db)
	adminRepo := repository.NewAdminRepo(db)
	mediaRepo := repository.NewMediaRepo(db)
	eventRepo := repository.NewEventRepo(db)

	// Handlers
	companyHandler := handler.NewCompanyHandler(companyRepo, serviceRepo)
	authHandler := handler.NewAuthHandler(authRepo, companyRepo, cfg.JWTSecret)
	quoteHandler := handler.NewQuoteHandler(quoteRepo, companyRepo, mailer)
	registrationHandler := handler.NewRegistrationHandler(registrationRepo)
	panelHandler := handler.NewPanelHandler(serviceRepo, quoteRepo, companyRepo)
	adminHandler := handler.NewAdminHandler(adminRepo, mailer, cfg.InitialPassword)
	healthHandler := handler.NewHealthHandler(db)
	mediaHandler := handler.NewMediaHandler(mediaRepo, companyRepo, store)
	analyticsHandler := handler.NewAnalyticsHandler(eventRepo, companyRepo)

	r := router.Setup(router.Deps{
		Company:      companyHandler,
		Auth:         authHandler,
		Quote:        quoteHandler,
		Registration: registrationHandler,
		Panel:        panelHandler,
		Admin:        adminHandler,
		Health:       healthHandler,
		Media:        mediaHandler,
		Analytics:    analyticsHandler,
		Storage:      store,
		StorageDir:   cfg.StorageDir,
		JWTSecret:    cfg.JWTSecret,
		CORSOrigin:   cfg.CORSOrigin,
	})

	slog.Info("server starting", "port", cfg.Port, "env", cfg.AppEnv)
	if err := r.Run(":" + cfg.Port); err != nil {
		slog.Error("server failed", "err", err)
	}
}
