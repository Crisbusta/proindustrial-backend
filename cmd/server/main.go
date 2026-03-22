package main

import (
	"log"

	"github.com/crisbusta/proindustrial-backend-public/internal/config"
	"github.com/crisbusta/proindustrial-backend-public/internal/database"
	"github.com/crisbusta/proindustrial-backend-public/internal/handler"
	"github.com/crisbusta/proindustrial-backend-public/internal/repository"
	"github.com/crisbusta/proindustrial-backend-public/internal/router"
)

func main() {
	cfg := config.Load()

	db := database.Connect(cfg.DatabaseURL)
	database.RunMigrations(cfg.DatabaseURL)

	// Repos
	companyRepo := repository.NewCompanyRepo(db)
	authRepo := repository.NewAuthRepo(db)
	quoteRepo := repository.NewQuoteRepo(db)
	serviceRepo := repository.NewServiceRepo(db)
	registrationRepo := repository.NewRegistrationRepo(db)

	// Handlers
	companyHandler := handler.NewCompanyHandler(companyRepo)
	authHandler := handler.NewAuthHandler(authRepo, companyRepo, cfg.JWTSecret)
	quoteHandler := handler.NewQuoteHandler(quoteRepo)
	registrationHandler := handler.NewRegistrationHandler(registrationRepo)
	panelHandler := handler.NewPanelHandler(serviceRepo, quoteRepo, companyRepo)

	r := router.Setup(router.Deps{
		Company:      companyHandler,
		Auth:         authHandler,
		Quote:        quoteHandler,
		Registration: registrationHandler,
		Panel:        panelHandler,
		JWTSecret:    cfg.JWTSecret,
		CORSOrigin:   cfg.CORSOrigin,
	})

	log.Printf("server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
