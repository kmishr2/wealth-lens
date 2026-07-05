package transactions

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/middleware"
)

const maxCSVImportBytes = 2 << 20

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

	var req TransactionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}

	transaction, err := h.service.Create(userID, portfolioID, req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusCreated, transaction)
}

func (h *Handler) ImportCSV(c *gin.Context) {
	userID, portfolioID, ok := parseUserAndPortfolio(c)
	if !ok {
		return
	}
	accountID, err := uuid.Parse(c.Param("accountId"))
	if err != nil {
		common.RespondError(c, common.NotFound("Account not found"))
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCSVImportBytes)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		common.RespondError(c, common.BadRequest("CSV file is required"))
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		common.RespondError(c, common.BadRequest("CSV file could not be opened"))
		return
	}
	defer file.Close()
	response, err := h.service.ImportCSV(userID, portfolioID, accountID, file)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	if response.RowsImported == 0 {
		common.RespondError(c, common.BadRequest("CSV file contains no transaction rows"))
		return
	}
	c.Header("X-Imported-Rows", fmt.Sprint(response.RowsImported))
	common.RespondOK(c, http.StatusCreated, response)
}

func (h *Handler) List(c *gin.Context) {
	userID, portfolioID, ok := parseUserAndPortfolio(c)
	if !ok {
		return
	}

	transactions, err := h.service.List(userID, portfolioID, common.ParsePagination(c))
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, transactions)
}

func (h *Handler) Get(c *gin.Context) {
	userID, portfolioID, transactionID, ok := parseUserPortfolioTransaction(c)
	if !ok {
		return
	}

	transaction, err := h.service.Get(userID, portfolioID, transactionID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, transaction)
}

func (h *Handler) Reverse(c *gin.Context) {
	userID, portfolioID, transactionID, ok := parseUserPortfolioTransaction(c)
	if !ok {
		return
	}

	var req TransactionReversalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}

	reversal, err := h.service.Reverse(userID, portfolioID, transactionID, req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusCreated, reversal)
}

func (h *Handler) Correct(c *gin.Context) {
	userID, portfolioID, transactionID, ok := parseUserPortfolioTransaction(c)
	if !ok {
		return
	}

	var req TransactionCorrectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}

	response, err := h.service.Correct(userID, portfolioID, transactionID, req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusCreated, response)
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

func parseUserPortfolioTransaction(c *gin.Context) (uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	userID, portfolioID, ok := parseUserAndPortfolio(c)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	transactionID, err := uuid.Parse(c.Param("transactionId"))
	if err != nil {
		common.RespondError(c, common.NotFound("Transaction not found"))
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}

	return userID, portfolioID, transactionID, true
}
