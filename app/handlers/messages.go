package handlers

import (
	"errors"
	"fmt"
	"messages/app/db"
	"messages/app/models"
	"messages/app/views/messages"
	"strconv"
	"time"

	v "github.com/anthdm/superkit/validate"

	"github.com/go-chi/chi/v5"

	"github.com/anthdm/superkit/kit"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func HandleMessagesList(kit *kit.Kit) error {
	data := &messages.IndexPageData{
		FormValues: getBaseMessageFormValues(),
	}

	messagesList, err := models.Messages().All(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}
	data.MessagesList = messagesList
	return kit.Render(messages.Index(data))
}

func HandleMessageGet(kit *kit.Kit) error {
	paramId := chi.URLParam(kit.Request, "id")

	messageId, err := strconv.ParseInt(paramId, 10, 64)
	if err != nil {
		return errors.New("Not found")
	}
	// // Get the message from the database
	dbMessage, err := models.FindMessage(kit.Request.Context(), db.Query, messageId)
	if err != nil {
		return err
	}

	data := &messages.PageMessageEditData{
		FormValues: getBaseMessageFormValues(),
		FormErrors: v.Errors{},
	}

	data.FormValues.DateRangeFrom = dbMessage.DisplayFrom.Format(time.RFC3339)
	data.FormValues.DateRangeTo = dbMessage.DisplayTo.Format(time.RFC3339)
	data.FormValues.Message = dbMessage.Message
	data.FormValues.Title = dbMessage.Title
	data.FormValues.Language = dbMessage.Language
	data.FormValues.ID = dbMessage.ID

	return kit.Render(messages.PageMessageEdit(data))
}

var createMessageSchema = v.Schema{
	"dateRangeFrom": v.Rules(v.Required),
	"dateRangeTo":   v.Rules(v.Required),
	"message":       v.Rules(v.Required),
	"title":         v.Rules(v.Required),
	"language":      v.Rules(v.Required),
}

func HandleMessageCreate(kit *kit.Kit) error {
	formValues := getBaseMessageFormValues()

	fmt.Println("Request: ", kit.Request)
	errors, ok := v.Request(kit.Request, &formValues, createMessageSchema)
	if !ok {
		return kit.Render(messages.MessageForm(formValues, errors))
	}

	displayFrom, err := time.Parse(time.RFC3339, formValues.DateRangeFrom)
	if err != nil {
		return err
	}

	displayTo, err := time.Parse(time.RFC3339, formValues.DateRangeTo)
	if err != nil {
		return err
	}

	message := models.Message{
		DisplayFrom: displayFrom,
		DisplayTo:   displayTo,
		Message:     formValues.Message,
		Title:       formValues.Title,
		Language:    formValues.Language,
		UserId:      1,
	}

	err = message.Insert(kit.Request.Context(), db.Query, boil.Infer())
	if err != nil {
		return err
	}

	return kit.Redirect(200, "/messages")
}

func HandleMessageUpdate(kit *kit.Kit) error {
	paramId := chi.URLParam(kit.Request, "id")

	messageId, err := strconv.ParseInt(paramId, 10, 64)
	if err != nil {
		return errors.New("Not found")
	}

	formValues := getBaseMessageFormValues()
	formValues.ID = messageId

	errors, ok := v.Request(kit.Request, &formValues, createMessageSchema)
	if !ok {
		return kit.Render(messages.MessageForm(formValues, errors))
	}

	displayFrom, err := time.Parse(time.RFC3339, formValues.DateRangeFrom)
	if err != nil {
		return err
	}

	displayTo, err := time.Parse(time.RFC3339, formValues.DateRangeTo)
	if err != nil {
		return err
	}

	message, err := models.Messages(
		models.MessageWhere.ID.EQ(formValues.ID),
	).One(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}

	message.Message = formValues.Message
	message.Title = formValues.Title
	message.Language = formValues.Language

	message.DisplayFrom = displayFrom
	message.DisplayTo = displayTo

	_, err = message.Update(kit.Request.Context(), db.Query, boil.Infer())
	if err != nil {
		return err
	}

	return kit.Redirect(200, "/messages")
}

func HandleMessageDelete(kit *kit.Kit) error {
	paramId := chi.URLParam(kit.Request, "id")

	messageId, err := strconv.ParseInt(paramId, 10, 64)
	if err != nil {
		return errors.New("Not found")
	}

	message, err := models.Messages(
		models.MessageWhere.ID.EQ(messageId),
	).One(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}

	_, err = message.Delete(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}

	return kit.Redirect(200, "/messages")
}

func getBaseMessageFormValues() messages.MessageFormValues {
	return messages.MessageFormValues{
		DateMin:       time.Now(),
		DateMax:       time.Now().AddDate(1, 0, 0),
		DateRangeFrom: "",
		DateRangeTo:   "",
		Message:       "",
		Title:         "",
	}
}
