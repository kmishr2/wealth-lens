package valuations

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/portfolios/:portfolioId/valuation", handler.GetCurrent)
}
