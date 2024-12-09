package websockets

import (
	"log"
	"messenger-pigeon-app/internal/model"
	"messenger-pigeon-app/pkg/repository"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	connectionManager = &ConnectionManager{}
	workerPool        = NewWorkerPool(10) // Pool com 10 workers
)

// WorkerPool gerencia o processamento de mensagens
type WorkerPool struct {
	jobQueue chan model.UserMessage
	wg       sync.WaitGroup
}

// Gerenciador de conexões WebSocket (thread-safe)
type ConnectionManager struct {
	connections sync.Map // Armazena conexões WebSocket por ID de usuário
}

func (cm *ConnectionManager) AddConnection(userID int64, conn *websocket.Conn) {
	cm.connections.Store(userID, conn)
}

func (cm *ConnectionManager) GetConnection(userID int64) (*websocket.Conn, bool) {
	conn, ok := cm.connections.Load(userID)
	if ok {
		return conn.(*websocket.Conn), true
	}
	return nil, false
}

func (cm *ConnectionManager) RemoveConnection(userID int) {
	cm.connections.Delete(userID)
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	pool := &WorkerPool{
		jobQueue: make(chan model.UserMessage, 100), // Buffer de 100 mensagens
	}
	for i := 0; i < numWorkers; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}
	return pool
}

func (pool *WorkerPool) worker() {
	defer pool.wg.Done()
	for job := range pool.jobQueue {
		processChatMessage(job)
	}
}

func (pool *WorkerPool) Submit(job model.UserMessage) {
	select {
	case pool.jobQueue <- job:
		// Mensagem adicionada à fila
	default:
		log.Println("Message queue is full. Dropping message:", job)
	}
}

func (pool *WorkerPool) Shutdown() {
	close(pool.jobQueue)
	pool.wg.Wait()
}

// Processa mensagens de forma individual
func processChatMessage(message model.UserMessage) {
	if conn, ok := connectionManager.GetConnection(int64(message.MessageTo)); ok {
		if err := conn.WriteJSON(message); err != nil {
			log.Println("Error sending message:", err)
		}
	} else {
		log.Println("Recipient is not connected:", message.MessageTo)
	}
}

// Lida com mensagens WebSocket em lotes para otimização
func handleWebSocketMessages() {
	batch := make([]model.UserMessage, 0, 10) // Lote de até 10 mensagens
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case message := <-workerPool.jobQueue:
			batch = append(batch, message)
			if len(batch) >= 10 {
				flushChatMessages(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				flushChatMessages(batch)
				batch = batch[:0]
			}
		}
	}
}

func flushChatMessages(batch []model.UserMessage) {
	messagesByRecipient := make(map[int64][]model.UserMessage)

	for _, message := range batch {
		messagesByRecipient[int64(message.MessageTo)] = append(messagesByRecipient[int64(message.MessageTo)], message)
	}

	for userID, messages := range messagesByRecipient {
		if conn, ok := connectionManager.GetConnection(userID); ok {
			if err := conn.WriteJSON(messages); err != nil {
				log.Printf("Error sending messages to user %d: %v", userID, err)
			}
		} else {
			log.Printf("User %d is not connected", userID)
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
