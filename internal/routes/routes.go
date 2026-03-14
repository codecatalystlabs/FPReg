package routes

import (
	"net/http"

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
}

func Setup(r *gin.Engine, h Handlers, authSvc *service.AuthService, auditSvc *service.AuditService) {
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/", func(c *gin.Context) { c.HTML(http.StatusOK, "login.html", nil) })
	r.GET("/dashboard", func(c *gin.Context) { c.HTML(http.StatusOK, "dashboard.html", nil) })
	r.GET("/register", func(c *gin.Context) { c.HTML(http.StatusOK, "register_form.html", nil) })
	r.GET("/submissions", func(c *gin.Context) { c.HTML(http.StatusOK, "submissions.html", nil) })
	r.GET("/submission/:id", func(c *gin.Context) { c.HTML(http.StatusOK, "submission_detail.html", nil) })
	r.GET("/guide", func(c *gin.Context) { c.HTML(http.StatusOK, "guide.html", nil) })
	r.GET("/admin/users", func(c *gin.Context) { c.HTML(http.StatusOK, "users.html", nil) })
	r.GET("/admin/facilities", func(c *gin.Context) { c.HTML(http.StatusOK, "facilities.html", nil) })
	r.GET("/admin/audit-logs", func(c *gin.Context) { c.HTML(http.StatusOK, "audit_logs.html", nil) })

	api := r.Group("/api/v1")
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
}
