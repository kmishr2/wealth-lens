package projections

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.POST("/portfolios/:portfolioId/projections/sip", handler.CalculateSIP)
}
