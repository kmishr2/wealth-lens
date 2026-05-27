package assets

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.POST("/assets", handler.Create)
	router.GET("/assets", handler.List)
	router.GET("/assets/:assetId", handler.Get)
}
