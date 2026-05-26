package accounts

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.POST("/portfolios/:portfolioId/accounts", handler.Create)
	router.GET("/portfolios/:portfolioId/accounts", handler.List)
	router.GET("/portfolios/:portfolioId/accounts/:accountId", handler.Get)
	router.PATCH("/portfolios/:portfolioId/accounts/:accountId", handler.Update)
	router.DELETE("/portfolios/:portfolioId/accounts/:accountId", handler.Delete)
}
