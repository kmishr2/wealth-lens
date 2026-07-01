package allocations

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetCurrent(c *gin.Context) {
	userID, portfolioID, ok := parseUserAndPortfolio(c)
	if !ok {
		return
	}

	allocation, err := h.service.GetCurrent(userID, portfolioID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, allocation)
}

func (h *Handler) GetConcentration(c *gin.Context) {
	userID, portfolioID, ok := parseUserAndPortfolio(c)
	if !ok {
		return
	}

	result, err := h.service.GetConcentration(userID, portfolioID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, result)
}

func (h *Handler) GetDiversificationAlerts(c *gin.Context) {
	userID, portfolioID, ok := parseUserAndPortfolio(c)
	if !ok {
		return
	}
	result, err := h.service.GetDiversificationAlerts(userID, portfolioID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, result)
}

func (h *Handler) CalculateRebalancing(c *gin.Context) {
	userID, portfolioID, ok := parseUserAndPortfolio(c)
	if !ok {
		return
	}

	var req RebalancingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}

	result, err := h.service.CalculateRebalancing(userID, portfolioID, req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, result)
}

func parseUserAndPortfolio(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		common.RespondError(c, err)
		return uuid.Nil, uuid.Nil, false
	}

	portfolioID, err := uuid.Parse(c.Param("portfolioId"))
	if err != nil {
		common.RespondError(c, common.NotFound("Portfolio not found"))
		return uuid.Nil, uuid.Nil, false
	}

	return userID, portfolioID, true
}
