package middleware

import (
	"fpreg/internal/service"
	"fpreg/internal/utils"

	"github.com/gin-gonic/gin"
)

func AuditLog(auditSvc *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("audit_service", auditSvc)
		c.Set("client_ip", utils.GetClientIP(c))
		c.Set("client_ua", c.GetHeader("User-Agent"))
		c.Next()
	}
}

func GetAuditInfo(c *gin.Context) (string, string) {
	ipVal, _ := c.Get("client_ip")
	uaVal, _ := c.Get("client_ua")
	ip, _ := ipVal.(string)
	ua, _ := uaVal.(string)
	return ip, ua
}
