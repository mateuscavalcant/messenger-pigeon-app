package routes

import (
	"messenger-pigeon-app/config/middleware"
	"messenger-pigeon-app/pkg/controllers"

	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.RouterGroup) {
	r.Use(middleware.AuthMiddleware())
	r.POST("/chat/:username", controllers.Chat)
	r.POST("/create-message/:username", controllers.CreateNewMessage)
	r.GET("/websocket/chat/:username", controllers.WebSocketChat)
	r.POST("/messages", controllers.Messages)
	r.GET("/websokcet/messages", controllers.WebSocketMessages)
}
