package controllers

import (
	"fmt"
	"log"
	"messenger-pigeon-app/pkg/model"
	"messenger-pigeon-app/repository"
	"messenger-pigeon-app/services"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// userConnections mapeia IDs de usuário para conexões WebSocket
// Canais para enviar mensagens
var (
	userMessageConnections map[int64]*websocket.Conn
	messageQueue           = make(chan model.UserMessage, 100) // Buffer de 100 mensagens

	connectionMessageMutexes sync.Map
)

func init() {
	userMessageConnections = make(map[int64]*websocket.Conn)
	go handleWebSocketMessages()

}

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
	partnerID, err := repository.MessageGetUserIDByUsername(username) // Supondo uma função no services
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID"})
		return
	}

	messages, err := services.GetChatMessages(id, partnerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
		return
	}

	chatPartnerName, chatPartnerIcon, err := services.GetChatPartnerInfo(partnerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chat partner info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages":    messages,
		"chatPartner": gin.H{"name": chatPartnerName, "iconBase64": chatPartnerIcon},
	})
}

// Função para enviar mensagens para o canal
func sendMessage(message model.UserMessage) {
	messageQueue <- message
}

// Função para lidar com as mensagens WebSocket
func handleWebSocketMessages() {
	for {
		// Aguarda mensagens no canal
		message := <-messageQueue

		// Verifique se o destinatário está conectado
		destConn, ok := userMessageConnections[int64(message.MessageTo)]
		if !ok {
			log.Println("Recipient is not connected")
			continue
		}

		// Envie a mensagem para o destinatário
		err := destConn.WriteJSON(message) // Use WriteJSON para enviar mensagens JSON via WebSocket
		if err != nil {
			log.Println("Error sending message:", err)
			continue
		}
	}
}

// WebSocketHandler é um manipulador HTTP para a rota WebSocket.
func WebSocketHandler(c *gin.Context) {
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
	// Atualizar a conexão para WebSocket
	ws, err := websocket.Upgrade(c.Writer, c.Request, nil, 1024, 1024)
	if err != nil {
		log.Println("Erro ao atualizar para WebSocket:", err)
		return
	}
	defer ws.Close()

	// Registre a conexão com o usuário
	userMessageConnections[int64(id)] = ws

	// Aguardar mensagens do usuário
	HandleMessages(ws)
}

// Enviar mensagens para o canal
func HandleMessages(ws *websocket.Conn) {
	defer ws.Close()

	for {
		var msg model.UserMessage
		err := ws.ReadJSON(&msg) // Use ReadJSON para ler mensagens JSON do WebSocket
		if err != nil {
			log.Println("Error receiving message:", err)
			return
		}

		// Envie a mensagem para o canal
		sendMessage(msg)
	}
}
