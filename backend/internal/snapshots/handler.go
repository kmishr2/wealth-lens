package snapshots

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
	userID, portfolioID, ok := parseUserAndPortfolio(c)
	if !ok {
		return
	}

	var req SnapshotCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}

	snapshot, err := h.service.CreateDaily(userID, portfolioID, req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusCreated, snapshot)
}

func (h *Handler) List(c *gin.Context) {
	userID, portfolioID, ok := parseUserAndPortfolio(c)
	if !ok {
		return
	}

	snapshots, err := h.service.List(userID, portfolioID, common.ParsePagination(c))
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, snapshots)
}

func (h *Handler) ListWeeklyPerformance(c *gin.Context) {
	userID, portfolioID, ok := parseUserAndPortfolio(c)
	if !ok {
		return
	}

	snapshots, err := h.service.ListWeeklyPerformance(userID, portfolioID, common.ParsePagination(c))
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, snapshots)
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
