package transactions

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.POST("/portfolios/:portfolioId/transactions", handler.Create)
	router.POST("/portfolios/:portfolioId/accounts/:accountId/transaction-imports", handler.ImportCSV)
	router.GET("/portfolios/:portfolioId/transactions", handler.List)
	router.GET("/portfolios/:portfolioId/transactions/:transactionId", handler.Get)
	router.POST("/portfolios/:portfolioId/transactions/:transactionId/reversals", handler.Reverse)
	router.POST("/portfolios/:portfolioId/transactions/:transactionId/corrections", handler.Correct)
}
