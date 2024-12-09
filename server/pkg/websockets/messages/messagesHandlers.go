package messages

import (
	"log"
	"messenger-pigeon-app/internal/model"
	"time"

	"github.com/gorilla/websocket"
)

var workerPool = NewWorkerPoolMessages(10)

// Função para lidar com as mensagens WebSocket de forma eficiente em lote
func HandleWebSocketMessagesHome() {
	batch := make([]model.UserMessage, 0, 10) // Processar lotes de 10 mensagens
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case message := <-workerPool.jobQueue:
			batch = append(batch, message)
			if len(batch) >= 10 {
				flushBatch(batch)
				batch = batch[:0] // Limpar o batch
			}
		case <-ticker.C:
			if len(batch) > 0 {
				flushBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// HandleMessages gerencia mensagens recebidas via WebSocket.
func HandleMessages(ws *websocket.Conn, userID int) {
	defer func() {
		ws.Close()
		connectionManager.RemoveConnection(int64(userID))
	}()

	ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.SetPongHandler(func(appData string) error {
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg model.UserMessage
		if err := ws.ReadJSON(&msg); err != nil {
			log.Println("Error receiving message:", err)
			return
		}

		// Envia a mensagem para processamento no pool.
		workerPool.Submit(msg)
	}
}
