package utils

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPagination(c *gin.Context) (page, perPage int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ = strconv.Atoi(c.DefaultQuery("per_page", "25"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 25
	}
	if perPage > 100 {
		perPage = 100
	}
	return
}

func CalcOffset(page, perPage int) int {
	return (page - 1) * perPage
}

func CalcTotalPages(total int64, perPage int) int {
	return int(math.Ceil(float64(total) / float64(perPage)))
}

func GetClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return ip
	}
	return c.ClientIP()
}
