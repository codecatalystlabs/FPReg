package utils

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPagination(c *gin.Context) (page, perPage int) {
	return GetPaginationOrMax(c, 25, 100)
}

// GetPaginationOrMax parses page and per_page; perPage is clamped to maxPerPage.
func GetPaginationOrMax(c *gin.Context, defaultPerPage, maxPerPage int) (page, perPage int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ = strconv.Atoi(c.DefaultQuery("per_page", strconv.Itoa(defaultPerPage)))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
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
