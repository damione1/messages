package handlers

import (
	"errors"
	"messages/app/db"
	"messages/app/models"

	messagesView "messages/app/views/messages"

	"github.com/anthdm/superkit/kit"
)

func HandleMessages(kit *kit.Kit) error {
	messagesList, err := models.Messages().All(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}
	return kit.Render(messagesView.Index(messagesList))
}
func HandleMessage(kit *kit.Kit) error {
	return errors.New("Not implemented")
}
