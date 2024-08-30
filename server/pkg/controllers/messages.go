package controllers

import (
	"fmt"
	"log"
	"messenger-pigeon-app/pkg/repository"
	"messenger-pigeon-app/pkg/services"
	"messenger-pigeon-app/pkg/websockets"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func Messages(c *gin.Context) {
	userId, exists := c.Get("id")
	if !exists {
		log.Println("User ID not found in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in session"})
		return
	}

	id, errId := strconv.Atoi(fmt.Sprintf("%v", userId))
	if errId != nil {
		log.Println("Error: ", errId)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		c.Abort()
		return
	}

	chats, err := services.GetUserChats(int64(id))
	if err != nil {
		log.Println("Error in service layer:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	currentUsername, err := repository.GetUsernameByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user Username"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"currentUsername": gin.H{"username": currentUsername},
		"chats":           chats,
	})
}

func WebSocketMessages(c *gin.Context) {
	ws, err := websocket.Upgrade(c.Writer, c.Request, nil, 1024, 1024)
	if err != nil {
		log.Println("Error: ", err)
		return
	}

	defer ws.Close()

	userID := websockets.GetUserIDFromContext(c)
	if userID == 0 {
		return
	}

	// Registrar a conexão
	websockets.UserConnectionsMessages[int64(userID)] = ws

	// Iniciar o controle de inatividade
	go websockets.StartInactivityTimerMessages(ws, userID)

	// Iniciar o manuseio de mensagens
	websockets.HandleMessages(ws, userID)
}
