package services

import (
	"fmt"
	"messenger-pigeon-app/config/database"
	"messenger-pigeon-app/internal/model"
	"messenger-pigeon-app/pkg/repository"
)

func GetUserChats(userID int64) ([]model.UserMessage, error) {
	db := database.GetDB()

	chats, err := repository.FetchUserChats(db, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user chats: %w", err)
	}

	return chats, nil
}
