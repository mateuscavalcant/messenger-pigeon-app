package chat

import (
	"log"
	"messenger-pigeon-app/internal/model"
)

func FlushChatMessages(batch []model.UserMessage) {
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
