package handlers

import (
	"messages/app/db"
	"messages/app/helpers"
	"messages/app/models"
	"time"

	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Message struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type Response struct {
	Origin   string    `json:"yourDomain"`
	Messages []Message `json:"messages"`
	Error    string    `json:"error,omitempty"`
}

func HandleApi(kit *kit.Kit) error {
	request := kit.Request
	kit.Response.Header().Set("Content-Type", "application/json")
	lang := chi.URLParam(kit.Request, "language")

	originDomain := request.Header.Get("Origin")
	response := Response{
		Origin:   originDomain,
		Messages: make([]Message, 0),
	}

	if !helpers.IsValidDomain(originDomain) {
		response.Error = "Invalid domain"
		kit.JSON(400, response)
		return nil
	}

	if !IsValidLanguage(lang) {
		response.Error = "Invalid language"
		kit.JSON(400, response)
		return nil
	}

	dbWebsite, err := models.Websites(
		models.WebsiteWhere.URL.EQ(originDomain),
	).One(kit.Request.Context(), db.Query)
	if err != nil {
		response.Error = "Unknown domain"
		kit.JSON(400, response)
		return nil
	}

	messagesIds, err := models.WebsitesMessages(
		models.WebsitesMessageWhere.WebsiteId.EQ(dbWebsite.ID),
	).All(kit.Request.Context(), db.Query)
	if err != nil {
		kit.JSON(200, response)
		return nil
	}

	messagesIdsList := make([]int64, 0, len(messagesIds))
	for _, message := range messagesIds {
		messagesIdsList = append(messagesIdsList, message.MessageId)
	}

	mod := []qm.QueryMod{
		models.MessageWhere.ID.IN(messagesIdsList),
		models.MessageWhere.Language.EQ(lang),
		models.MessageWhere.DisplayTo.GT(time.Now()),
	}

	if !dbWebsite.Staging {
		mod = append(mod, models.MessageWhere.DisplayFrom.LT(time.Now()))
	}

	dbMessageList, err := models.Messages(mod...).All(kit.Request.Context(), db.Query)

	response.Messages = make([]Message, 0, len(dbMessageList))
	for _, dbMessage := range dbMessageList {
		message := Message{
			Title:   dbMessage.Title,
			Message: dbMessage.Message,
		}
		if dbWebsite.Staging && dbMessage.DisplayFrom.After(time.Now()) {
			message.Message = "[Preview] " + message.Message
		}

		response.Messages = append(response.Messages, message)
	}
	kit.JSON(200, response)
	return nil
}

func IsValidLanguage(lang string) bool {
	return lang == "en" || lang == "fr"
}
