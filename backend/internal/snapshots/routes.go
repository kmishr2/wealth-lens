package snapshots

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.POST("/portfolios/:portfolioId/snapshots", handler.Create)
	router.GET("/portfolios/:portfolioId/snapshots", handler.List)
}
