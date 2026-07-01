package goals

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.POST("/portfolios/:portfolioId/goals", handler.Create)
	router.GET("/portfolios/:portfolioId/goals", handler.List)
	router.PATCH("/portfolios/:portfolioId/goals/:goalId", handler.Update)
	router.DELETE("/portfolios/:portfolioId/goals/:goalId", handler.Delete)
	router.POST("/portfolios/:portfolioId/goals/:goalId/monthly-snapshots", handler.CreateMonthlySnapshot)
	router.GET("/portfolios/:portfolioId/goals/:goalId/monthly-snapshots", handler.ListMonthlySnapshots)
}
