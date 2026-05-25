package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}

	response, err := h.service.Register(req)
	if err != nil {
		common.RespondError(c, err)
		return
	}

	common.RespondOK(c, http.StatusCreated, response)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}

	response, err := h.service.Login(req)
	if err != nil {
		common.RespondError(c, err)
		return
	}

	common.RespondOK(c, http.StatusOK, response)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}

	response, err := h.service.Refresh(req.RefreshToken)
	if err != nil {
		common.RespondError(c, err)
		return
	}

	common.RespondOK(c, http.StatusOK, response)
}

func (h *Handler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.Logout(req.RefreshToken); err != nil {
		common.RespondError(c, err)
		return
	}

	common.RespondNoContent(c)
}
