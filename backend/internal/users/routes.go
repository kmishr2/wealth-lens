package users

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/users/me", handler.GetMe)
	router.PATCH("/users/me", handler.UpdateMe)
}
