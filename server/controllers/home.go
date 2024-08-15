package controllers

import (
	"fmt"
	"log"
	"messenger-pigeon-app/pkg/model"
	"messenger-pigeon-app/services"
	"net/http"
	"strconv"
	"sync"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// userConnections mapeia IDs de usuário para conexões WebSocket
// Canais para enviar mensagens
var (
	userChatsConnections   map[int64]*websocket.Conn
	chatsQueue             = make(chan model.UserMessage, 100) // Buffer de 100 mensagens
	connectionChatsMutexes sync.Map
)

func init() {
	userChatsConnections = make(map[int64]*websocket.Conn)
	go handleWebSocketChats()
}

func FeedChats(c *gin.Context) {
	userId, exists := c.Get("id")
	if !exists {
		log.Println("ID do usuário não encontrado na sessão")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID do usuário não encontrado na sessão"})
		return
	}

	id, errId := strconv.Atoi(fmt.Sprintf("%v", userId))
	if errId != nil {
		log.Println("Erro ao converter ID do usuário para int:", errId)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID do usuário inválido"})
		c.Abort()
		return
	}

	chats, err := services.GetUserChats(int64(id))
	if err != nil {
		log.Println("Error in services layer:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Verificar se o usuário possui uma conexão WebSocket ativa
	conn, ok := userChatsConnections[int64(id)]
	if ok {
		connMutex, _ := connectionChatsMutexes.LoadOrStore(int64(id), &sync.Mutex{})
		connMutex.(*sync.Mutex).Lock()
		defer connMutex.(*sync.Mutex).Unlock()

		// Enviar as mensagens via WebSocket
		err = conn.WriteJSON(gin.H{"chats": chats})
		if err != nil {
			log.Println("Error sending messages via WebSocket:", err)
		}
	} else {
		log.Println("User does not have an active WebSocket connection")
	}

	c.JSON(http.StatusOK, gin.H{
		"chats": chats,
	})
}

func WebSocketFeedChats(c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		log.Println("Token JWT não fornecido")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token JWT não fornecido"})
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Método de assinatura inesperado: %v", token.Header["alg"])
		}
		return []byte("your_secret_key"), nil
	})

	if err != nil || !token.Valid {
		log.Println("Token JWT inválido:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token JWT inválido"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		log.Println("Token JWT inválido")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token JWT inválido"})
		return
	}

	userId, ok := claims["id"].(float64)
	if !ok {
		log.Println("ID do usuário não encontrado no token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID do usuário não encontrado no token"})
		return
	}

	ws, err := websocket.Upgrade(c.Writer, c.Request, nil, 1024, 1024)
	if err != nil {
		log.Println("Erro ao atualizar para WebSocket:", err)
		return
	}
	defer ws.Close()

	userChatsConnections[int64(userId)] = ws
	HandleMessagesChats(ws)
}

// Função para enviar mensagens para o canal
func sendMessageChats(message model.UserMessage) {
	chatsQueue <- message
}

// Função para lidar com as mensagens WebSocket
func handleWebSocketChats() {
	for {
		// Aguarda mensagens no canal
		message := <-chatsQueue

		// Verifique se o destinatário está conectado
		destConn, ok := userChatsConnections[int64(message.MessageTo)]
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

// Enviar mensagens para o canal
func HandleMessagesChats(ws *websocket.Conn) {
	defer ws.Close()

	for {
		var msg model.UserMessage
		err := ws.ReadJSON(&msg) // Use ReadJSON para ler mensagens JSON do WebSocket
		if err != nil {
			log.Println("Error receiving message:", err)
			return
		}

		// Envie a mensagem para o canal
		sendMessageChats(msg)
	}
}
