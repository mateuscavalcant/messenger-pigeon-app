package repository

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"messenger-pigeon-app/pkg/model"
)

func FetchUserChats(db *sql.DB, userID int64) ([]model.UserMessage, error) {
	query := `
    SELECT 
        user.id AS user_id, 
        user.username, 
        user.name, 
        user.icon, 
        user_message.content, 
        user_message.created_at
    FROM user_message
    JOIN user ON (user.id = user_message.messageTo OR user.id = user_message.messageBy)
    WHERE (user_message.messageBy = ? OR user_message.messageTo = ?)
    AND user.id != ?
    AND user_message.created_at = (
        SELECT MAX(user_message2.created_at)
        FROM user_message AS user_message2
        WHERE (
            (user_message2.messageBy = user_message.messageBy AND user_message2.messageTo = user_message.messageTo) 
            OR (user_message2.messageBy = user_message.messageTo AND user_message2.messageTo = user_message.messageBy)
        )
    )
    ORDER BY user_message.created_at DESC
    `

	rows, err := db.Query(query, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query statements: %w", err)
	}
	defer rows.Close()

	var chats []model.UserMessage
	for rows.Next() {
		var chat model.UserMessage
		var icon []byte
		var createdAtString string

		err := rows.Scan(&chat.UserID, &chat.CreatedBy, &chat.Name, &icon, &chat.Content, &createdAtString)
		if err != nil {
			return nil, fmt.Errorf("failed to scan statement: %w", err)
		}

		var imageBase64 string
		if icon != nil {
			imageBase64 = base64.StdEncoding.EncodeToString(icon)
		}

		chats = append(chats, model.UserMessage{
			UserID:     chat.UserID,
			CreatedBy:  chat.CreatedBy,
			Name:       chat.Name,
			IconBase64: imageBase64,
			Content:    chat.Content,
			CreatedAt:  chat.CreatedAt,
		})
	}

	return chats, nil
}
