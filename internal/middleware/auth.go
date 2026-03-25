package middleware

import (
	"strings"

	"fpreg/internal/models"
	"fpreg/internal/service"
	"fpreg/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuthRequired(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			utils.RespondUnauthorized(c, "Missing authorization header")
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.RespondUnauthorized(c, "Invalid authorization format")
			return
		}

		claims, err := authSvc.ValidateAccessToken(parts[1])
		if err != nil {
			utils.RespondUnauthorized(c, "Invalid or expired token")
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		if claims.FacilityID != nil {
			c.Set("facility_id", *claims.FacilityID)
		}
		if strings.TrimSpace(claims.District) != "" {
			c.Set("user_district", strings.TrimSpace(claims.District))
		}
		c.Next()
	}
}

func RoleRequired(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("user_role")
		if !exists {
			utils.RespondForbidden(c)
			return
		}
		userRole := roleVal.(models.Role)
		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}
		utils.RespondForbidden(c)
	}
}

func FacilityScoped() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("user_role")
		if role == models.RoleSuperAdmin || role == models.RoleDistrictBiostatistician {
			c.Next()
			return
		}
		facilityID, exists := c.Get("facility_id")
		if !exists {
			utils.RespondForbidden(c)
			return
		}
		c.Set("scoped_facility_id", facilityID.(uuid.UUID))
		c.Next()
	}
}

func GetUserID(c *gin.Context) uuid.UUID {
	v, _ := c.Get("user_id")
	return v.(uuid.UUID)
}

func GetFacilityID(c *gin.Context) *uuid.UUID {
	v, exists := c.Get("facility_id")
	if !exists {
		return nil
	}
	id := v.(uuid.UUID)
	return &id
}

func GetScopedFacilityID(c *gin.Context) *uuid.UUID {
	role, _ := c.Get("user_role")
	if role == models.RoleSuperAdmin {
		if qf := c.Query("facility_id"); qf != "" {
			id, err := uuid.Parse(qf)
			if err == nil {
				return &id
			}
		}
		return nil
	}
	if role == models.RoleDistrictBiostatistician {
		return nil
	}
	v, exists := c.Get("facility_id")
	if !exists {
		return nil
	}
	id := v.(uuid.UUID)
	return &id
}

// GetUserDistrict returns the district scope for district_biostatistician (from JWT claims).
func GetUserDistrict(c *gin.Context) string {
	v, ok := c.Get("user_district")
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return strings.TrimSpace(s)
}
