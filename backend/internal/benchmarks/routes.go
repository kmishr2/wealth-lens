package benchmarks

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.POST("/benchmarks", handler.Create)
	router.GET("/benchmarks", handler.List)
	router.POST("/benchmarks/:benchmarkId/observations", handler.CreateObservation)
	router.GET("/benchmarks/:benchmarkId/observations", handler.ListObservations)
	router.GET("/portfolios/:portfolioId/benchmarks/:benchmarkId/comparison", handler.ComparePortfolio)
}
