package accounts

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

	var req AccountCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}

	account, err := h.service.Create(userID, portfolioID, req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusCreated, account)
}

func (h *Handler) List(c *gin.Context) {
	userID, portfolioID, ok := parseUserAndPortfolio(c)
	if !ok {
		return
	}

	accounts, err := h.service.List(userID, portfolioID, common.ParsePagination(c))
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, accounts)
}

func (h *Handler) Get(c *gin.Context) {
	userID, portfolioID, accountID, ok := parseUserPortfolioAccount(c)
	if !ok {
		return
	}

	account, err := h.service.Get(userID, portfolioID, accountID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, account)
}

func (h *Handler) Update(c *gin.Context) {
	userID, portfolioID, accountID, ok := parseUserPortfolioAccount(c)
	if !ok {
		return
	}

	var req AccountUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}

	account, err := h.service.Update(userID, portfolioID, accountID, req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, account)
}

func (h *Handler) Delete(c *gin.Context) {
	userID, portfolioID, accountID, ok := parseUserPortfolioAccount(c)
	if !ok {
		return
	}

	if err := h.service.Delete(userID, portfolioID, accountID); err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondNoContent(c)
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

func parseUserPortfolioAccount(c *gin.Context) (uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	userID, portfolioID, ok := parseUserAndPortfolio(c)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	accountID, err := uuid.Parse(c.Param("accountId"))
	if err != nil {
		common.RespondError(c, common.NotFound("Account not found"))
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	return userID, portfolioID, accountID, true
}
