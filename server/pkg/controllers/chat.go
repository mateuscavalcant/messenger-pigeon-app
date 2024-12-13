package controllers

import (
	"fmt"
	"log"
	"messenger-pigeon-app/internal/err"
	"messenger-pigeon-app/internal/model"
	"messenger-pigeon-app/pkg/repository"
	"messenger-pigeon-app/pkg/services"
	"messenger-pigeon-app/pkg/websockets/chat"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Chat é um manipulador HTTP que lida com solicitações de chat.
func Chat(c *gin.Context) {
	userId, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in session"})
		return
	}

	id, err := strconv.Atoi(fmt.Sprintf("%v", userId))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	username := c.Param("username")
	partnerID, err := repository.MessageGetUserIDByUsername(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID"})
		return
	}

	var req model.Pagination

	if err := c.ShouldBindJSON(&req); err != nil {
		req.LastMessageID = 0
	}

	log.Println("new limit: ", req.NewLimit)

	req.Limit = 20

	if req.NewLimit > 20 {
		req.Limit = req.NewLimit
	}

	log.Println("limit: ", req.Limit)

	messages, err := services.GetChatMessages(id, partnerID, req.LastMessageID, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
		return
	}

	userInfosName, userInfosUsername, userInfosIcon, err := services.GetChatInfos(partnerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chat partner info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages":      messages,
		"limitMessages": gin.H{"limit": req.Limit},
		"userInfos":     gin.H{"name": userInfosName, "username": userInfosUsername, "iconBase64": userInfosIcon},
	})
}

// WebSocketChat é um manipulador HTTP para a rota websockets.
func WebSocketChat(c *gin.Context) {
	var conn chat.ConnectionManager
	ws, err := websocket.Upgrade(c.Writer, c.Request, nil, 1024, 1024)
	if err != nil {
		log.Println("Error:", err)
		return
	}
	defer ws.Close()

	userID := GetUserIDFromContext(c)
	if userID == 0 {
		return
	}

	// Registrar a conexão
	conn.AddConnection(int64(userID), ws)

	// Iniciar o manuseio de mensagens
	chat.HandleChatMessages(ws, userID)
}

func CreateNewMessage(c *gin.Context) {
	var errResp err.ErrorResponse

	// Parse do corpo da requisição
	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := c.Param("username")
	content := strings.TrimSpace(c.PostForm("content"))
	userId, exists := c.Get("id")
	if !exists {
		log.Println("User ID not found in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in session"})
		return
	}

	id, err := strconv.Atoi(fmt.Sprintf("%v", userId))
	if err != nil {
		log.Println("Error: ", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Validação básica
	if content == "" {
		errResp.Error["content"] = "Values are missing!"
	}
	if len(errResp.Error) > 0 {
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	// Chama o service para enviar a mensagem
	messageID, err := chat.SendChatMessage(id, username, content)
	if err != nil {
		log.Println("Error sending message:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	resp := map[string]interface{}{
		"messageID": messageID,
		"message":   "Message sent successfully",
	}

	c.JSON(http.StatusOK, resp)
}

// Helper para extrair o ID do usuário do contexto
func GetUserIDFromContext(c *gin.Context) int {
	userId, exists := c.Get("id")
	if !exists {
		log.Println("User ID not found in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in session"})
		return 0
	}

	var id int
	idFloat, ok := userId.(float64)
	if !ok {
		id = int(idFloat)

	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
	}

	return id
}
