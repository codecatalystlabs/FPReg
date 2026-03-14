package handler

import (
	"net/http"

	"fpreg/internal/repository"
	"fpreg/internal/utils"

	"github.com/gin-gonic/gin"
)

type OptionSetHandler struct {
	repo *repository.OptionSetRepository
}

func NewOptionSetHandler(repo *repository.OptionSetRepository) *OptionSetHandler {
	return &OptionSetHandler{repo: repo}
}

// ListOptionSets godoc
// @Summary      List all option sets grouped by category
// @Tags         option-sets
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/option-sets [get]
func (h *OptionSetHandler) ListGrouped(c *gin.Context) {
	grouped, err := h.repo.FindAllGrouped()
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to load option sets")
		return
	}
	utils.RespondOK(c, grouped)
}

// ListByCategory godoc
// @Summary      List option sets by category
// @Tags         option-sets
// @Produce      json
// @Security     BearerAuth
// @Param        category path string true "Category name"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/option-sets/{category} [get]
func (h *OptionSetHandler) ListByCategory(c *gin.Context) {
	category := c.Param("category")
	items, err := h.repo.FindByCategory(category)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to load option set")
		return
	}
	utils.RespondOK(c, items)
}

// ListCategories godoc
// @Summary      List all option set category names
// @Tags         option-sets
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/option-sets/categories [get]
func (h *OptionSetHandler) ListCategories(c *gin.Context) {
	cats, err := h.repo.Categories()
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to load categories")
		return
	}
	utils.RespondOK(c, cats)
}
