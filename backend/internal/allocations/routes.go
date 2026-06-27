package allocations

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/portfolios/:portfolioId/allocation", handler.GetCurrent)
	router.POST("/portfolios/:portfolioId/rebalancing", handler.CalculateRebalancing)
}
