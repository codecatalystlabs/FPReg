package routes

import (
	"log"
	"net/http"
	"path/filepath"

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

	webRoot := config.WebAssetsRoot()
	if webRoot != "" {
		staticDir := filepath.Join(webRoot, "web", "static")
		tplPattern := filepath.Join(webRoot, "web", "templates", "*")
		base.Static("/static", staticDir)
		r.LoadHTMLGlob(tplPattern)

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
		log.Println("Web templates not found (no web/templates under cwd or above the executable), running in API-only mode")
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

		facilityAdmin := facilities.Group("")
		facilityAdmin.Use(middleware.RoleRequired(models.RoleSuperAdmin))
		{
			// Static path must be registered before /:id so "districts" is not parsed as UUID.
			facilityAdmin.GET("/districts", h.Facility.ListDistricts)
			facilityAdmin.POST("", h.Facility.Create)
			facilityAdmin.PUT("/:id", h.Facility.Update)
			facilityAdmin.DELETE("/:id", h.Facility.Delete)
		}

		facilities.GET("/:id", h.Facility.GetByID)
	}

	// Users
	users := api.Group("/users")
	users.Use(middleware.AuthRequired(authSvc))
	users.Use(middleware.RoleRequired(models.RoleSuperAdmin, models.RoleFacilityAdmin, models.RoleDistrictBiostatistician))
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
	regWrite := middleware.RoleRequired(
		models.RoleSuperAdmin, models.RoleFacilityAdmin, models.RoleFacilityUser,
		models.RoleDistrictBiostatistician,
	)
	{
		registrations.GET("", h.Registration.List)
		registrations.GET("/:id", h.Registration.GetByID)
		registrations.POST("", regWrite, h.Registration.Create)
		registrations.PUT("/:id", regWrite, h.Registration.Update)
		registrations.DELETE("/:id", middleware.RoleRequired(
			models.RoleSuperAdmin, models.RoleFacilityAdmin,
		), h.Registration.Delete)
	}

	// Audit logs (scoped: district biostatistician sees own district only)
	auditLogs := api.Group("/audit-logs")
	auditLogs.Use(middleware.AuthRequired(authSvc))
	auditLogs.Use(middleware.RoleRequired(models.RoleSuperAdmin, models.RoleFacilityAdmin, models.RoleDistrictBiostatistician))
	{
		auditLogs.GET("", h.Audit.List)
	}

	// FP monthly report: all authenticated facility roles may view; DHIS2 POST limited to admins / district biostat.
	reports := api.Group("/reports/family-planning")
	reports.Use(middleware.AuthRequired(authSvc))
	{
		reportReaders := middleware.RoleRequired(
			models.RoleSuperAdmin, models.RoleFacilityAdmin, models.RoleFacilityUser,
			models.RoleReviewer, models.RoleDistrictBiostatistician,
		)
		reports.GET("/monthly", reportReaders, h.FPReport.Monthly)
		reports.GET("/payload-preview", reportReaders, h.FPReport.PayloadPreview)
		reports.POST("/sync", middleware.RoleRequired(
			models.RoleSuperAdmin, models.RoleFacilityAdmin, models.RoleDistrictBiostatistician,
		), h.FPReport.Sync)
	}
}
