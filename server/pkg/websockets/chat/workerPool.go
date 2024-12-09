package chat

import (
	"log"
	"messenger-pigeon-app/internal/model"
	"sync"
)

// WorkerPool gerencia o processamento de mensagens
type WorkerPool struct {
	jobQueue chan model.UserMessage
	wg       sync.WaitGroup
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
		ProcessChatMessage(job)
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
func ProcessChatMessage(message model.UserMessage) {
	if conn, ok := connectionManager.GetConnection(int64(message.MessageTo)); ok {
		if err := conn.WriteJSON(message); err != nil {
			log.Println("Error sending message:", err)
		}
	} else {
		log.Println("Recipient is not connected:", message.MessageTo)
	}
}
