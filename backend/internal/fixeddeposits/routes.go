package fixeddeposits

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.POST("/portfolios/:portfolioId/accounts/:accountId/fixed-deposits", handler.Create)
	router.GET("/portfolios/:portfolioId/accounts/:accountId/fixed-deposits", handler.List)
	router.POST("/portfolios/:portfolioId/accounts/:accountId/fixed-deposits/:fixedDepositId/values", handler.CreateValue)
}
