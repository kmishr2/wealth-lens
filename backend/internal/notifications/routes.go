package notifications

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/notifications", handler.List)
}
