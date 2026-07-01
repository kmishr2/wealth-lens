package beta

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/portfolios/:portfolioId/benchmarks/:benchmarkId/beta", handler.Get)
}
