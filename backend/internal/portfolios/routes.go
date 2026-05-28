package portfolios

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.POST("/portfolios", handler.Create)
	router.GET("/portfolios", handler.List)
	router.GET("/portfolios/:portfolioId", handler.Get)
	router.PATCH("/portfolios/:portfolioId", handler.Update)
	router.DELETE("/portfolios/:portfolioId", handler.Delete)
}
