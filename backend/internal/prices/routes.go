package prices

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.POST("/assets/:assetId/prices", handler.Create)
	router.GET("/assets/:assetId/prices", handler.List)
	router.GET("/assets/:assetId/prices/latest", handler.GetLatest)
}
