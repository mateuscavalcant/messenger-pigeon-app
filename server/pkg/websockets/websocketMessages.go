package websockets

import (
	"log"
	"messenger-pigeon-app/internal/model"

	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Mapeamento de conexões WebSocket por ID de usuário
var (
	UserConnectionsMessages map[int64]*websocket.Conn
	workerPoolMessages      *WorkerPoolMessages
)

func init() {
	UserConnectionsMessages = make(map[int64]*websocket.Conn)
	workerPoolMessages = NewworkerPoolMessages(10) // Pool com 10 workers
	go handleWebSocketMessagesHome()
}

// Pool de workers para processar mensagens
type WorkerPoolMessages struct {
	workers  int
	jobQueue chan model.UserMessage
	wg       sync.WaitGroup
}

func NewworkerPoolMessages(numWorkers int) *WorkerPoolMessages {
	pool := &WorkerPoolMessages{
		workers:  numWorkers,
		jobQueue: make(chan model.UserMessage, 100), // Buffer com 100 mensagens
	}
	pool.startWorkers()
	return pool
}

func (pool *WorkerPoolMessages) startWorkers() {
	for i := 0; i < pool.workers; i++ {
		pool.wg.Add(1)
		go func() {
			defer pool.wg.Done()
			for job := range pool.jobQueue {
				processMessages(job)
			}
		}()
	}
}

func (pool *WorkerPoolMessages) Submit(job model.UserMessage) {
	select {
	case pool.jobQueue <- job:
		// Mensagem enviada para o pool com sucesso
	default:
		// Buffer de mensagens cheio, mensagem será descartada ou reprocessada.
	}
}

func (pool *WorkerPoolMessages) Shutdown() {
	close(pool.jobQueue)
	pool.wg.Wait()
}

func processMessages(message model.UserMessage) {
	// Processar a mensagem e enviar via WebSocket
	if conn, ok := UserConnectionsMessages[int64(message.MessageTo)]; ok {
		err := conn.WriteJSON(message)
		if err != nil {
			log.Println("Error sending message:", err)
		}
	} else {
		log.Println("Recipient is not connected")
	}
}

// Função para enviar mensagens para o pool
func sendMessages(message model.UserMessage) {
	workerPoolMessages.Submit(message)
}

// Função para lidar com as mensagens WebSocket de forma eficiente em lote
func handleWebSocketMessagesHome() {
	batch := make([]model.UserMessage, 0, 10) // Processar lotes de 10 mensagens
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case message := <-workerPoolMessages.jobQueue:
			batch = append(batch, message)
			if len(batch) >= 10 {
				flushMessages(batch)
				batch = batch[:0] // Limpar o batch
			}
		case <-ticker.C:
			if len(batch) > 0 {
				flushMessages(batch)
				batch = batch[:0]
			}
		}
	}
}

func flushMessages(batch []model.UserMessage) {
	for _, message := range batch {
		if conn, ok := UserConnectionsMessages[int64(message.MessageTo)]; ok {
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
func StartInactivityTimerMessages(ws *websocket.Conn, userID int) {
	inactivityDuration := 30 * time.Second
	inactivityTimer := time.NewTimer(inactivityDuration)

	for {
		select {
		case <-inactivityTimer.C:
			// Fechar a conexão após 30 segundos de inatividade
			log.Println("Closing connection due to inactivity:", userID)
			ws.Close()
			delete(UserConnectionsMessages, int64(userID))
			return
		case <-time.After(1 * time.Second): // Checa a cada segundo se a conexão ainda está ativa
			if _, isConnected := UserConnectionsMessages[int64(userID)]; !isConnected {
				inactivityTimer.Stop()
				return
			}
		}
	}
}

// Função para gerenciar o timeout de conexão ociosa com o PongHandler e redefinir o timer de inatividade
func HandleMessages(ws *websocket.Conn, userID int) {
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
		go ResetInactivityTimerMessages(userID)

		sendMessages(msg)
	}
}

// Função para redefinir o timer de inatividade
func ResetInactivityTimerMessages(userID int) {
	if _, ok := UserConnectionsMessages[int64(userID)]; ok {
		log.Println("Resetting user idle timer:", userID)
	}
}
