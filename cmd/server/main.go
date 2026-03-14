package main

import (
	"log"
	"os"

	"fpreg/docs"
	"fpreg/internal/config"
	"fpreg/internal/database"
	"fpreg/internal/handler"
	"fpreg/internal/middleware"
	"fpreg/internal/repository"
	"fpreg/internal/routes"
	"fpreg/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	docs.SwaggerInfo.BasePath = cfg.BasePath + "/api/v1"
	docs.SwaggerInfo.Host = ""

	gin.SetMode(cfg.GinMode)

	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("Warning: could not create logs directory: %v", err)
	}

	db := database.Connect(cfg)
	database.Migrate(db)
	database.Seed(db, cfg.SeedAdminEmail, cfg.SeedAdminPassword, cfg.SeedAdminName)

	// Repositories
	userRepo := repository.NewUserRepository(db)
	facilityRepo := repository.NewFacilityRepository(db)
	optionSetRepo := repository.NewOptionSetRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	clientNumberRepo := repository.NewClientNumberRepository(db)
	registrationRepo := repository.NewRegistrationRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// Services
	auditSvc := service.NewAuditService(auditRepo)
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, cfg)
	userSvc := service.NewUserService(userRepo, auditSvc)
	facilitySvc := service.NewFacilityService(facilityRepo, auditSvc)
	registrationSvc := service.NewRegistrationService(registrationRepo, clientNumberRepo, facilityRepo, auditSvc)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc, auditSvc)
	userHandler := handler.NewUserHandler(userSvc)
	facilityHandler := handler.NewFacilityHandler(facilitySvc)
	optionSetHandler := handler.NewOptionSetHandler(optionSetRepo)
	registrationHandler := handler.NewRegistrationHandler(registrationSvc)
	auditHandler := handler.NewAuditHandler(auditSvc)

	r := gin.Default()
	r.Use(middleware.CORS())

	routes.Setup(r, routes.Handlers{
		Auth:         authHandler,
		User:         userHandler,
		Facility:     facilityHandler,
		OptionSet:    optionSetHandler,
		Registration: registrationHandler,
		Audit:        auditHandler,
	}, authSvc, auditSvc, cfg)

	port := cfg.AppPort
	log.Printf("Starting HMIS FP Register server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
