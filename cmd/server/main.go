package main

import (
	"log"

	"github.com/crisbusta/proindustrial-backend-public/internal/config"
	"github.com/crisbusta/proindustrial-backend-public/internal/database"
	"github.com/crisbusta/proindustrial-backend-public/internal/handler"
	"github.com/crisbusta/proindustrial-backend-public/internal/notify"
	"github.com/crisbusta/proindustrial-backend-public/internal/repository"
	"github.com/crisbusta/proindustrial-backend-public/internal/router"
)

func main() {
	cfg := config.Load()
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

	// Handlers
	companyHandler := handler.NewCompanyHandler(companyRepo, serviceRepo)
	authHandler := handler.NewAuthHandler(authRepo, companyRepo, cfg.JWTSecret)
	quoteHandler := handler.NewQuoteHandler(quoteRepo, companyRepo, mailer)
	registrationHandler := handler.NewRegistrationHandler(registrationRepo)
	panelHandler := handler.NewPanelHandler(serviceRepo, quoteRepo, companyRepo)
	adminHandler := handler.NewAdminHandler(adminRepo, mailer, cfg.InitialPassword)

	r := router.Setup(router.Deps{
		Company:      companyHandler,
		Auth:         authHandler,
		Quote:        quoteHandler,
		Registration: registrationHandler,
		Panel:        panelHandler,
		Admin:        adminHandler,
		JWTSecret:    cfg.JWTSecret,
		CORSOrigin:   cfg.CORSOrigin,
	})

	log.Printf("server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
