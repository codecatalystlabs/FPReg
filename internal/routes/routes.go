package routes

import (
	"log"
	"net/http"
	"os"

	"fpreg/internal/config"
	"fpreg/internal/handler"
	"fpreg/internal/middleware"
	"fpreg/internal/models"
	"fpreg/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handlers struct {
	Auth         *handler.AuthHandler
	User         *handler.UserHandler
	Facility     *handler.FacilityHandler
	OptionSet    *handler.OptionSetHandler
	Registration *handler.RegistrationHandler
	Audit        *handler.AuditHandler
	FPReport     *handler.FPReportHandler
}

func Setup(r *gin.Engine, h Handlers, authSvc *service.AuthService, auditSvc *service.AuditService, cfg *config.Config) {
	if cfg.BasePath != "" {
		r.GET(cfg.BasePath, func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, cfg.BasePath+"/") })
	}
	base := r.Group(cfg.BasePath)

	base.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL(cfg.BasePath+"/swagger/doc.json"),
	))

	if _, err := os.Stat("web/templates"); err == nil {
		base.Static("/static", "./web/static")
		r.LoadHTMLGlob("web/templates/*")

		tplData := gin.H{"BasePath": cfg.BasePath}
		base.GET("/", func(c *gin.Context) { c.HTML(http.StatusOK, "login.html", tplData) })
		base.GET("/dashboard", func(c *gin.Context) { c.HTML(http.StatusOK, "dashboard.html", tplData) })
		base.GET("/register", func(c *gin.Context) { c.HTML(http.StatusOK, "register_form.html", tplData) })
		base.GET("/submissions", func(c *gin.Context) { c.HTML(http.StatusOK, "submissions.html", tplData) })
		base.GET("/submission/:id", func(c *gin.Context) { c.HTML(http.StatusOK, "submission_detail.html", tplData) })
		base.GET("/guide", func(c *gin.Context) { c.HTML(http.StatusOK, "guide.html", tplData) })
		base.GET("/admin/users", func(c *gin.Context) { c.HTML(http.StatusOK, "users.html", tplData) })
		base.GET("/admin/facilities", func(c *gin.Context) { c.HTML(http.StatusOK, "facilities.html", tplData) })
		base.GET("/admin/audit-logs", func(c *gin.Context) { c.HTML(http.StatusOK, "audit_logs.html", tplData) })
		base.GET("/reports/fp-monthly", func(c *gin.Context) { c.HTML(http.StatusOK, "fp_monthly_report.html", tplData) })
	} else {
		log.Println("Web templates not found, running in API-only mode")
	}

	api := base.Group("/api/v1")
	api.Use(middleware.CORS())
	api.Use(middleware.AuditLog(auditSvc))

	// Public auth endpoints
	auth := api.Group("/auth")
	{
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.Refresh)
	}

	// Protected auth endpoints
	authProtected := api.Group("/auth")
	authProtected.Use(middleware.AuthRequired(authSvc))
	{
		authProtected.POST("/logout", h.Auth.Logout)
		authProtected.GET("/me", h.Auth.Me)
		authProtected.POST("/change-password", h.Auth.ChangePassword)
	}

	// Option sets (authenticated)
	optionSets := api.Group("/option-sets")
	optionSets.Use(middleware.AuthRequired(authSvc))
	{
		optionSets.GET("", h.OptionSet.ListGrouped)
		optionSets.GET("/categories", h.OptionSet.ListCategories)
		optionSets.GET("/:category", h.OptionSet.ListByCategory)
	}

	// Facilities
	facilities := api.Group("/facilities")
	facilities.Use(middleware.AuthRequired(authSvc))
	{
		facilities.GET("", h.Facility.List)
		facilities.GET("/:id", h.Facility.GetByID)

		facilityAdmin := facilities.Group("")
		facilityAdmin.Use(middleware.RoleRequired(models.RoleSuperAdmin))
		{
			facilityAdmin.POST("", h.Facility.Create)
			facilityAdmin.PUT("/:id", h.Facility.Update)
			facilityAdmin.DELETE("/:id", h.Facility.Delete)
		}
	}

	// Users
	users := api.Group("/users")
	users.Use(middleware.AuthRequired(authSvc))
	users.Use(middleware.RoleRequired(models.RoleSuperAdmin, models.RoleFacilityAdmin))
	{
		users.GET("", h.User.List)
		users.GET("/:id", h.User.GetByID)
		users.POST("", h.User.Create)
		users.PUT("/:id", h.User.Update)
		users.PATCH("/:id/deactivate", h.User.Deactivate)
		users.POST("/import", h.User.Import)
	}

	// Registrations
	registrations := api.Group("/registrations")
	registrations.Use(middleware.AuthRequired(authSvc))
	registrations.Use(middleware.FacilityScoped())
	{
		registrations.GET("", h.Registration.List)
		registrations.GET("/:id", h.Registration.GetByID)
		registrations.POST("", h.Registration.Create)
		registrations.PUT("/:id", h.Registration.Update)
		registrations.DELETE("/:id", middleware.RoleRequired(
			models.RoleSuperAdmin, models.RoleFacilityAdmin,
		), h.Registration.Delete)
	}

	// Audit logs (superadmin / facility_admin only)
	auditLogs := api.Group("/audit-logs")
	auditLogs.Use(middleware.AuthRequired(authSvc))
	auditLogs.Use(middleware.RoleRequired(models.RoleSuperAdmin, models.RoleFacilityAdmin))
	{
		auditLogs.GET("", h.Audit.List)
	}

	// FP Monthly Reports (superadmin / facility_admin / district_biostatistician)
	reports := api.Group("/reports/family-planning")
	reports.Use(middleware.AuthRequired(authSvc))
	reports.Use(middleware.RoleRequired(models.RoleSuperAdmin, models.RoleFacilityAdmin, models.RoleReviewer))
	{
		reports.GET("/monthly", h.FPReport.Monthly)
		reports.GET("/payload-preview", h.FPReport.PayloadPreview)
		reports.POST("/sync", h.FPReport.Sync)
	}
}
