package handler

import (
	"net/http"

	"fpreg/internal/middleware"
	"fpreg/internal/models"
	"fpreg/internal/service"
	"fpreg/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc  *service.AuthService
	auditSvc *service.AuditService
}

func NewAuthHandler(authSvc *service.AuthService, auditSvc *service.AuditService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, auditSvc: auditSvc}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login godoc
// @Summary      Log in
// @Description  Authenticate user with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body LoginRequest true "Login credentials"
// @Success      200  {object} utils.APIResponse
// @Failure      401  {object} utils.APIError
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Email and password are required")
		return
	}

	ip := utils.GetClientIP(c)
	ua := c.GetHeader("User-Agent")

	pair, user, err := h.authSvc.Login(req.Email, req.Password, ip, ua)
	if err != nil {
		h.auditSvc.Log(nil, nil, models.AuditLoginFailed, "auth", "", ip, ua, "Failed login: "+req.Email)
		utils.RespondUnauthorized(c, err.Error())
		return
	}

	h.auditSvc.Log(&user.ID, user.FacilityID, models.AuditLogin, "auth", user.ID.String(), ip, ua, "")

	c.JSON(http.StatusOK, utils.APIResponse{
		Success: true,
		Data: gin.H{
			"tokens": pair,
			"user": gin.H{
				"id":          user.ID,
				"email":       user.Email,
				"full_name":   user.FullName,
				"role":        user.Role,
				"facility_id": user.FacilityID,
				"facility":    user.Facility,
			},
		},
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// Refresh godoc
// @Summary      Refresh tokens
// @Description  Exchange a refresh token for a new access + refresh token pair
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RefreshRequest true "Refresh token"
// @Success      200  {object} utils.APIResponse
// @Failure      401  {object} utils.APIError
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Refresh token is required")
		return
	}

	ip := utils.GetClientIP(c)
	ua := c.GetHeader("User-Agent")

	pair, err := h.authSvc.RefreshTokens(req.RefreshToken, ip, ua)
	if err != nil {
		h.auditSvc.Log(nil, nil, models.AuditTokenRefresh, "auth", "", ip, ua, "Failed refresh")
		utils.RespondUnauthorized(c, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{"tokens": pair})
}

// Logout godoc
// @Summary      Log out
// @Description  Revoke the refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RefreshRequest true "Refresh token to revoke"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondMessage(c, "Logged out")
		return
	}

	ip := utils.GetClientIP(c)
	ua := c.GetHeader("User-Agent")
	userID := middleware.GetUserID(c)

	_ = h.authSvc.Logout(req.RefreshToken)
	h.auditSvc.Log(&userID, nil, models.AuditLogout, "auth", userID.String(), ip, ua, "")
	utils.RespondMessage(c, "Logged out")
}

// Me godoc
// @Summary      Get current user
// @Description  Returns the currently authenticated user's profile
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	email, _ := c.Get("user_email")
	role, _ := c.Get("user_role")
	facilityID := middleware.GetFacilityID(c)

	utils.RespondOK(c, gin.H{
		"id":          userID,
		"email":       email,
		"role":        role,
		"facility_id": facilityID,
	})
}

// ChangePassword godoc
// @Summary      Change own password
// @Description  Allows the currently authenticated user to change their password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body ChangePasswordRequest true "Password change payload"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "current_password and new_password are required")
		return
	}
	userID := middleware.GetUserID(c)
	ip := utils.GetClientIP(c)
	ua := c.GetHeader("User-Agent")

	if err := h.authSvc.ChangePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	h.auditSvc.Log(&userID, nil, models.AuditUpdate, "auth", userID.String(), ip, ua, "Changed password")
	utils.RespondMessage(c, "Password changed successfully")
}
