package handlers

import (
	"context"
	"errors"
	"fmt"
	"messages/app/db"
	"messages/app/models"
	component_multiSelectField "messages/app/views/components/multiSelectField"
	"messages/app/views/messages"
	"messages/plugins/auth"
	"strconv"
	"time"

	v "github.com/anthdm/superkit/validate"

	"github.com/go-chi/chi/v5"

	"github.com/anthdm/superkit/kit"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func HandleMessagesList(kit *kit.Kit) error {
	data := &messages.IndexPageData{
		FormValues:   &messages.MessageFormValues{},
		FormSettings: getBaseMessageFormSettings(kit.Request.Context()),
	}

	dbMessagesList, err := models.Messages(
		qm.OrderBy("display_from DESC"),
	).All(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}

	messagesList := make([]*messages.MessageListItem, 0, len(dbMessagesList))
	for _, dbMessage := range dbMessagesList {
		messagesList = append(messagesList, &messages.MessageListItem{
			ID:          dbMessage.ID,
			Title:       dbMessage.Title,
			DisplayFrom: dbMessage.DisplayFrom,
			DisplayTo:   dbMessage.DisplayTo,
			Type:        dbMessage.Type,
			Language:    dbMessage.Language,
			Status:      getMessageStatus(dbMessage),
		})
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
	dbMessage, err := models.Messages(
		models.MessageWhere.ID.EQ(messageId),
	).One(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}

	dbWebsites, err := models.WebsitesMessages(
		models.WebsitesMessageWhere.MessageId.EQ(messageId),
	).All(kit.Request.Context(), db.Query)

	websites := make([]string, 0, len(dbWebsites))
	for _, website := range dbWebsites {
		websites = append(websites, fmt.Sprintf("%d", website.WebsiteId))
	}

	data := &messages.PageMessageEditData{
		FormValues: &messages.MessageFormValues{
			ID:            messageId,
			DateRangeFrom: dbMessage.DisplayFrom.Format(time.RFC3339),
			DateRangeTo:   dbMessage.DisplayTo.Format(time.RFC3339),
			Message:       dbMessage.Message,
			Title:         dbMessage.Title,
			Type:          dbMessage.Type,
			Language:      dbMessage.Language,
			Websites:      websites,
		},
		FormSettings: getBaseMessageFormSettings(kit.Request.Context()),
		FormErrors:   v.Errors{},
	}

	return kit.Render(messages.PageMessageEdit(data))
}

var createMessageSchema = v.Schema{
	"dateRangeFrom": v.Rules(v.Required),
	"dateRangeTo":   v.Rules(v.Required),
	"message":       v.Rules(v.Required),
	"title":         v.Rules(v.Required),
	"type":          v.Rules(v.Required),
	"language":      v.Rules(v.Required),
	"websites":      v.Rules(),
}

func HandleMessageCreate(kit *kit.Kit) error {
	auth := kit.Auth().(auth.Auth)
	formValues := &messages.MessageFormValues{}
	formSettings := getBaseMessageFormSettings(kit.Request.Context())

	errors, ok := v.Request(kit.Request, formValues, createMessageSchema)
	if !ok {
		return kit.Render(messages.MessageForm(formValues, formSettings, errors))
	}

	if err := component_multiSelectField.ParseMultiSelectFields(kit.Request, formValues); err != nil {
		// Handle error if multi-select parsing fails
		errors.Add("_error", err.Error())
		return kit.Render(messages.MessageForm(formValues, formSettings, errors))
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
		Type:        formValues.Type,
		Language:    formValues.Language,
		UserId:      int64(auth.UserID),
	}

	err = message.Insert(kit.Request.Context(), db.Query, boil.Infer())
	if err != nil {
		return err
	}

	err = upsertMessageWebsites(kit.Request.Context(), &message, formValues.Websites)
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

	formValues := &messages.MessageFormValues{
		ID: messageId,
	}
	formSettings := getBaseMessageFormSettings(kit.Request.Context())

	errors, ok := v.Request(kit.Request, formValues, createMessageSchema)

	err = component_multiSelectField.ParseMultiSelectFields(kit.Request, formValues)
	if err != nil || !ok {
		return kit.Render(messages.MessageForm(formValues, formSettings, errors))
	}

	displayFrom, err := time.Parse(time.RFC3339, formValues.DateRangeFrom)
	if err != nil {
		return err
	}

	displayTo, err := time.Parse(time.RFC3339, formValues.DateRangeTo)
	if err != nil {
		return err
	}

	dbmessage, err := models.Messages(
		models.MessageWhere.ID.EQ(formValues.ID),
	).One(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}

	dbmessage.Message = formValues.Message
	dbmessage.Title = formValues.Title
	dbmessage.Type = formValues.Type
	dbmessage.Language = formValues.Language

	dbmessage.DisplayFrom = displayFrom
	dbmessage.DisplayTo = displayTo

	_, err = dbmessage.Update(kit.Request.Context(), db.Query, boil.Infer())
	if err != nil {
		return err
	}

	err = upsertMessageWebsites(kit.Request.Context(), dbmessage, formValues.Websites)
	if err != nil {
		return err
	}

	return kit.Redirect(200, "/messages")
}

func upsertMessageWebsites(ctx context.Context, message *models.Message, websites []string) error {
	_, err := models.WebsitesMessages(
		models.WebsitesMessageWhere.MessageId.EQ(message.ID),
	).DeleteAll(ctx, db.Query)
	if err != nil {
		return err
	}

	for _, websiteId := range websites {
		websiteIdInt, err := strconv.ParseInt(websiteId, 10, 64)
		if err != nil {
			return err
		}

		websiteMessage := models.WebsitesMessage{
			WebsiteId: websiteIdInt,
			MessageId: message.ID,
		}

		err = websiteMessage.Insert(ctx, db.Query, boil.Infer())
		if err != nil {
			return err
		}
	}

	return nil
}

func HandleMessageDelete(kit *kit.Kit) error {
	paramId := chi.URLParam(kit.Request, "id")

	messageId, err := strconv.ParseInt(paramId, 10, 64)
	if err != nil {
		return errors.New("Not found")
	}

	_, err = models.Messages(
		models.MessageWhere.ID.EQ(messageId),
	).DeleteAll(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}

	_, err = models.WebsitesMessages(
		models.WebsitesMessageWhere.MessageId.EQ(messageId),
	).DeleteAll(kit.Request.Context(), db.Query)

	return kit.Redirect(200, "/messages")
}

func getBaseMessageFormSettings(ctx context.Context) *messages.MessageFormSettings {

	settings := &messages.MessageFormSettings{
		DateMin: time.Now(),
		DateMax: time.Now().AddDate(1, 0, 0),
	}

	dbWebsitesList, err := models.Websites().All(ctx, db.Query)
	if err != nil {
		return settings
	}

	settings.Websites = make(map[string]string, len(dbWebsitesList))
	for _, website := range dbWebsitesList {
		settings.Websites[fmt.Sprintf("%d", website.ID)] = fmt.Sprintf("%s (%s)", website.Name, website.URL)
	}

	return settings
}

func getMessageStatus(message *models.Message) string {
	switch {
	case message.DisplayFrom.After(time.Now()):
		return messages.ScheduledEnum
	case message.DisplayTo.Before(time.Now()):
		return messages.ExpiredEnum
	default:
		return messages.ActiveEnum
	}
}
