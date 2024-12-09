package chat

import (
	"log"
	"messenger-pigeon-app/internal/model"
	"messenger-pigeon-app/pkg/repository"
	"time"

	"github.com/gorilla/websocket"
)

var workerPool = NewWorkerPool(10)

// Lida com mensagens WebSocket em lotes para otimização
func HandleWebSocketMessages() {
	batch := make([]model.UserMessage, 0, 10) // Lote de até 10 mensagens
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case message := <-workerPool.jobQueue:
			batch = append(batch, message)
			if len(batch) >= 10 {
				FlushChatMessages(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				FlushChatMessages(batch)
				batch = batch[:0]
			}
		}
	}
}

// Gerencia mensagens WebSocket e reconexões
func HandleChatMessages(ws *websocket.Conn, userID int) {
	defer func() {
		ws.Close()
		connectionManager.RemoveConnection(userID)
	}()

	connectionManager.AddConnection(int64(userID), ws)
	ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.SetPongHandler(func(appData string) error {
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg model.UserMessage
		if err := ws.ReadJSON(&msg); err != nil {
			log.Println("Error receiving message:", err)
			break
		}
		workerPool.Submit(msg)
	}
}

// Envia uma mensagem (armazenando no banco de dados, se necessário)
func SendChatMessage(senderID int, receiverUsername, content string) (int64, error) {
	receiverID, err := repository.MessageGetUserIDByUsername(receiverUsername)
	if err != nil {
		return 0, err
	}

	message := model.UserMessage{
		MessageBy: senderID,
		MessageTo: receiverID,
		Content:   content,
	}

	messageID, err := repository.SaveMessage(message)
	if err != nil {
		return 0, err
	}

	if conn, isOnline := connectionManager.GetConnection(int64(receiverID)); isOnline {
		go func() {
			if err := conn.WriteJSON(message); err != nil {
				log.Printf("Error sending message to user %d: %v", receiverID, err)
			}
		}()
	} else {
		log.Printf("Recipient %d is offline. Message saved to database.", receiverID)
	}

	return messageID, nil
}
