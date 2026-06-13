package risk

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/portfolios/:portfolioId/risk", handler.Get)
}
