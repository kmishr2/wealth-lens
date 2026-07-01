package goals

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

func (h *Handler) Create(c *gin.Context) {
	userID, portfolioID, ok := parseUserPortfolio(c)
	if !ok {
		return
	}
	var req GoalCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}
	response, err := h.service.Create(userID, portfolioID, req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusCreated, response)
}

func (h *Handler) List(c *gin.Context) {
	userID, portfolioID, ok := parseUserPortfolio(c)
	if !ok {
		return
	}
	response, err := h.service.List(userID, portfolioID, common.ParsePagination(c))
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, response)
}

func (h *Handler) Update(c *gin.Context) {
	userID, portfolioID, goalID, ok := parseUserPortfolioGoal(c)
	if !ok {
		return
	}
	var req GoalUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}
	response, err := h.service.Update(userID, portfolioID, goalID, req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, response)
}

func (h *Handler) Delete(c *gin.Context) {
	userID, portfolioID, goalID, ok := parseUserPortfolioGoal(c)
	if !ok {
		return
	}
	if err := h.service.Delete(userID, portfolioID, goalID); err != nil {
		common.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) CreateMonthlySnapshot(c *gin.Context) {
	userID, portfolioID, goalID, ok := parseUserPortfolioGoal(c)
	if !ok {
		return
	}
	var req struct {
		SnapshotMonthEnd string `json:"snapshot_month_end"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}
	response, err := h.service.CreateMonthlySnapshot(userID, portfolioID, goalID, req.SnapshotMonthEnd)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusCreated, response)
}

func (h *Handler) ListMonthlySnapshots(c *gin.Context) {
	userID, portfolioID, goalID, ok := parseUserPortfolioGoal(c)
	if !ok {
		return
	}
	response, err := h.service.ListMonthlySnapshots(userID, portfolioID, goalID, common.ParsePagination(c))
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, response)
}

func parseUserPortfolio(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
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

func parseUserPortfolioGoal(c *gin.Context) (uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	userID, portfolioID, ok := parseUserPortfolio(c)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	goalID, err := uuid.Parse(c.Param("goalId"))
	if err != nil {
		common.RespondError(c, common.NotFound("Goal not found"))
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	return userID, portfolioID, goalID, true
}
