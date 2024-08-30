package websockets

import (
	"log"
	"messenger-pigeon-app/internal/model"
	"messenger-pigeon-app/pkg/repository"
	"net/http"

	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Mapeamento de conexões WebSocket por ID de usuário
var (
	UserConnections map[int64]*websocket.Conn
	workerPool      *WorkerPool
)

func init() {
	UserConnections = make(map[int64]*websocket.Conn)
	workerPool = NewWorkerPool(10) // Pool com 10 workers
	go handleWebSocketMessages()
}

// Pool de workers para processar mensagens
type WorkerPool struct {
	workers  int
	jobQueue chan model.UserMessage
	wg       sync.WaitGroup
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	pool := &WorkerPool{
		workers:  numWorkers,
		jobQueue: make(chan model.UserMessage, 100), // Buffer com 100 mensagens
	}
	pool.startWorkers()
	return pool
}

func (pool *WorkerPool) startWorkers() {
	for i := 0; i < pool.workers; i++ {
		pool.wg.Add(1)
		go func() {
			defer pool.wg.Done()
			for job := range pool.jobQueue {
				processChatMessage(job)
			}
		}()
	}
}

func (pool *WorkerPool) Submit(job model.UserMessage) {
	select {
	case pool.jobQueue <- job:
		// Mensagem enviada para o pool com sucesso
	default:
		// Buffer de mensagens cheio, mensagem será descartada ou reprocessada.
		log.Println()
	}
}

func (pool *WorkerPool) Shutdown() {
	close(pool.jobQueue)
	pool.wg.Wait()
}

func processChatMessage(message model.UserMessage) {
	// Processar a mensagem e enviar via WebSocket
	if conn, ok := UserConnections[int64(message.MessageTo)]; ok {
		err := conn.WriteJSON(message)
		if err != nil {
			log.Println("Error sending message:", err)
		}
	} else {
		log.Println("Recipient is not connected")
	}
}

// Função para enviar mensagens para o pool
func sendChatMessage(message model.UserMessage) {
	workerPool.Submit(message)
}

// Função para lidar com as mensagens WebSocket de forma eficiente em lote
func handleWebSocketMessages() {
	batch := make([]model.UserMessage, 0, 10) // Processar lotes de 10 mensagens
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case message := <-workerPool.jobQueue:
			batch = append(batch, message)
			if len(batch) >= 10 {
				flushChatMessages(batch)
				batch = batch[:0] // Limpar o batch
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
	for _, message := range batch {
		if conn, ok := UserConnections[int64(message.MessageTo)]; ok {
			// Enviar todas as mensagens em um único payload JSON
			err := conn.WriteJSON(batch)
			if err != nil {
				log.Println("Error sending messages:", err)
			}
		} else {
			log.Println("Recipient is not connected")
		}
	}
}

// Função para iniciar o controle de inatividade
func StartInactivityTimer(ws *websocket.Conn, userID int) {
	inactivityDuration := 30 * time.Second
	inactivityTimer := time.NewTimer(inactivityDuration)

	for {
		select {
		case <-inactivityTimer.C:
			// Fechar a conexão após 30 segundos de inatividade
			log.Println("Closing connection due to inactivity:", userID)
			ws.Close()
			delete(UserConnections, int64(userID))
			return
		case <-time.After(1 * time.Second): // Checa a cada segundo se a conexão ainda está ativa
			if _, isConnected := UserConnections[int64(userID)]; !isConnected {
				inactivityTimer.Stop()
				return
			}
		}
	}
}

// Função para gerenciar o timeout de conexão ociosa com o PongHandler e redefinir o timer de inatividade
func HandleChatMessages(ws *websocket.Conn, userID int) {
	defer ws.Close()

	ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.SetPongHandler(func(appData string) error {
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg model.UserMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Println("Error receiving message:", err)
			return
		}

		// Redefinir o timer de inatividade sempre que uma mensagem for recebida
		go ResetInactivityTimer(userID)

		sendChatMessage(msg)
	}
}

// Função para redefinir o timer de inatividade
func ResetInactivityTimer(userID int) {
	if _, ok := UserConnections[int64(userID)]; ok {
		log.Println("Resetting user idle timer:", userID)
	}
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

func SendChatMessage(senderID int, receiverUsername, content string) (int64, error) {
	// Obtém o ID do usuário destinatário
	receiverID, err := repository.MessageGetUserIDByUsername(receiverUsername)
	if err != nil {
		log.Println("Error getting recipient ID:", err)
		return 0, err
	}

	// Cria a mensagem
	message := model.UserMessage{
		MessageBy: senderID,
		MessageTo: receiverID,
		Content:   content,
	}

	// Salva a mensagem no banco de dados
	messageID, err := repository.SaveMessage(message)
	if err != nil {
		log.Println("Error saving message", err)
		return 0, err
	}

	// Verifica se o destinatário está online
	conn, isOnline := UserConnections[int64(receiverID)]
	if isOnline {
		// Envia a mensagem via WebSocket
		go func() {
			if err := conn.WriteJSON(message); err != nil {
				log.Printf("Error sending message via WebSocket to user %d: %v", receiverID, err)
			}
		}()
	} else {
		log.Printf("The %d recipient is not online. The message was only stored in the database.", receiverID)
	}

	return messageID, nil
}
