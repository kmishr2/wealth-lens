package fixeddeposits

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
	userID, portfolioID, accountID, ok := parseScope(c)
	if !ok {
		return
	}
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}
	response, err := h.service.Create(userID, portfolioID, accountID, req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusCreated, response)
}

func (h *Handler) List(c *gin.Context) {
	userID, portfolioID, accountID, ok := parseScope(c)
	if !ok {
		return
	}
	response, err := h.service.List(userID, portfolioID, accountID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, response)
}

func (h *Handler) CreateValue(c *gin.Context) {
	userID, portfolioID, accountID, ok := parseScope(c)
	if !ok {
		return
	}
	fixedDepositID, err := uuid.Parse(c.Param("fixedDepositId"))
	if err != nil {
		common.RespondError(c, common.NotFound("Fixed deposit not found"))
		return
	}
	var req ValueCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}
	response, err := h.service.CreateValue(userID, portfolioID, accountID, fixedDepositID, req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusCreated, response)
}

func parseScope(c *gin.Context) (uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		common.RespondError(c, err)
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	portfolioID, err := uuid.Parse(c.Param("portfolioId"))
	if err != nil {
		common.RespondError(c, common.NotFound("Portfolio not found"))
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	accountID, err := uuid.Parse(c.Param("accountId"))
	if err != nil {
		common.RespondError(c, common.NotFound("Account not found"))
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	return userID, portfolioID, accountID, true
}
