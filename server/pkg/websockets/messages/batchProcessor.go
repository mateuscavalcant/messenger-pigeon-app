package messages

import (
	"log"
	"messenger-pigeon-app/internal/model"
)

// flushBatch envia mensagens em lote para os usuários.
func flushBatch(batch []model.UserMessage) {
	messagesByUser := make(map[int64][]model.UserMessage)

	// Agrupa mensagens por destinatário.
	for _, msg := range batch {
		messagesByUser[int64(msg.MessageTo)] = append(messagesByUser[int64(msg.MessageTo)], msg)
	}

	// Envia os lotes para cada usuário.
	for userID, messages := range messagesByUser {
		conn, ok := connectionManager.GetConnection(userID)
		if !ok {
			log.Printf("User %d is not connected", userID)
			continue
		}

		if err := conn.WriteJSON(messages); err != nil {
			log.Printf("Error sending messages to user %d: %v", userID, err)
		}
	}
}
