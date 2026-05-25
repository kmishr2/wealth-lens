package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetMe(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		common.RespondError(c, err)
		return
	}

	user, err := h.service.GetMe(userID)
	if err != nil {
		common.RespondError(c, err)
		return
	}

	common.RespondOK(c, http.StatusOK, user)
}

func (h *Handler) UpdateMe(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		common.RespondError(c, err)
		return
	}

	var req UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}

	user, err := h.service.UpdateMe(userID, req)
	if err != nil {
		common.RespondError(c, err)
		return
	}

	common.RespondOK(c, http.StatusOK, user)
}
