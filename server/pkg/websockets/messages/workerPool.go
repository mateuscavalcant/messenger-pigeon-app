package messages

import (
	"log"
	"messenger-pigeon-app/internal/model"
	"sync"
)

// WorkerPoolMessages gerencia workers para processar mensagens.
type WorkerPoolMessages struct {
	workers  int
	jobQueue chan model.UserMessage
	wg       sync.WaitGroup
}

// NewWorkerPoolMessages cria uma nova instância do pool de workers.
func NewWorkerPoolMessages(numWorkers int) *WorkerPoolMessages {
	pool := &WorkerPoolMessages{
		workers:  numWorkers,
		jobQueue: make(chan model.UserMessage, 100),
	}
	pool.startWorkers()
	return pool
}

// startWorkers inicia os workers para processar mensagens.
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

// Submit adiciona um trabalho ao pool.
func (pool *WorkerPoolMessages) Submit(job model.UserMessage) {
	select {
	case pool.jobQueue <- job:
		// Trabalho enviado com sucesso.
	default:
		log.Println("Job queue full, message discarded")
	}
}

// Shutdown encerra o pool de workers.
func (pool *WorkerPoolMessages) Shutdown() {
	close(pool.jobQueue)
	pool.wg.Wait()
}

// processMessages processa uma mensagem.
func processMessages(message model.UserMessage) {
	// Processa a mensagem e a envia via WebSocket.
	conn, ok := connectionManager.GetConnection(int64(message.MessageTo))
	if !ok {
		log.Println("Recipient is not connected")
		return
	}

	if err := conn.WriteJSON(message); err != nil {
		log.Println("Error sending message:", err)
	}
}
