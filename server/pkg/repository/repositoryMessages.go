package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"messenger-pigeon-app/config/database"
	"messenger-pigeon-app/internal/model"
)

func MessageGetUserIDByUsername(username string) (int, error) {
	db := database.GetDB()
	var id int
	err := db.QueryRow("SELECT id FROM user WHERE username = ?", username).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("User not found")
		}
		log.Println("Error querying user ID:", err)
		return 0, err
	}
	return id, nil
}

// Obter mensagens entre usuários
func GetUserMessages(user1ID, user2ID int) ([]model.UserMessage, error) {
	db := database.GetDB()
	stmt, err := db.Prepare(`
		SELECT user_message.message_id, user_message.messageBy, user_message.content,
		       user.id, user.username, user.name, user.icon, user_message.created_at
		FROM user_message
		JOIN user ON user.id = user_message.messageBy
		WHERE (user_message.messageBy = ? AND user_message.messageTo = ?) OR 
		      (user_message.messageBy = ? AND user_message.messageTo = ?)
		ORDER BY user_message.created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(user1ID, user2ID, user2ID, user1ID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var messages []model.UserMessage
	for rows.Next() {
		var message model.UserMessage
		if err := rows.Scan(&message.MessageID, &message.MessageUserID, &message.Content, &message.UserID, &message.CreatedBy, &message.Name, &message.Icon, &message.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan rows: %w", err)
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func GetUsernameByID(userID int) (string, error) {
	db := database.GetDB()
	var username string
	err := db.QueryRow("SELECT username FROM user WHERE id = ?", userID).Scan(&username)
	if err != nil {
		log.Println("Erro ao consultar username:", err)
		return "", err
	}
	return username, nil
}
