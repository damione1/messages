package handlers

import (
	"context"
	"fmt"
	"messages/app/db"
	"messages/app/helpers"
	"messages/app/models"
	"messages/app/views/messages"
	"messages/plugins/auth"
	"net/http"
	"reflect"
	"strconv"
	"time"

	v "github.com/anthdm/superkit/validate"

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
	messageId, err := helpers.GetIdFromUrl(kit)
	if err != nil {
		return err
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
	"type":          v.Rules(v.Required, v.In([]string{"info", "warning", "danger"})),
	"language":      v.Rules(v.Required, v.In([]string{"en", "fr"})),
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

	if err := parseMultiSelectFields(kit.Request, formValues); err != nil {
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

	dbMessage := &models.Message{
		DisplayFrom: displayFrom,
		DisplayTo:   displayTo,
		Message:     formValues.Message,
		Title:       formValues.Title,
		Type:        formValues.Type,
		Language:    formValues.Language,
		UserId:      int64(auth.UserID),
	}

	err = dbMessage.Insert(kit.Request.Context(), db.Query, boil.Infer())
	if err != nil {
		return err
	}

	err = upsertMessageWebsites(kit.Request.Context(), dbMessage.ID, formValues.Websites)
	if err != nil {
		return err
	}

	return kit.Redirect(200, "/messages")
}

func HandleMessageUpdate(kit *kit.Kit) error {
	messageId, err := helpers.GetIdFromUrl(kit)
	if err != nil {
		return err
	}

	formValues := &messages.MessageFormValues{
		ID: messageId,
	}
	errors, ok := v.Request(kit.Request, formValues, createMessageSchema)

	formSettings := getBaseMessageFormSettings(kit.Request.Context())

	err = parseMultiSelectFields(kit.Request, formValues)
	if err != nil || !ok {
		return kit.Render(messages.MessageForm(formValues, formSettings, errors))
	}

	displayFrom, err := time.Parse(time.RFC3339, formValues.DateRangeFrom)
	if err != nil {
		errors.Add("form", "Failed to parse date range from")
		return kit.Render(messages.MessageForm(formValues, formSettings, errors))
	}

	displayTo, err := time.Parse(time.RFC3339, formValues.DateRangeTo)
	if err != nil {
		errors.Add("form", "Failed to parse date range to")
		return kit.Render(messages.MessageForm(formValues, formSettings, errors))
	}

	_, err = models.Messages(
		models.MessageWhere.ID.EQ(messageId),
	).UpdateAll(kit.Request.Context(), db.Query, models.M{
		models.MessageColumns.DisplayFrom: displayFrom,
		models.MessageColumns.DisplayTo:   displayTo,
		models.MessageColumns.Message:     formValues.Message,
		models.MessageColumns.Title:       formValues.Title,
		models.MessageColumns.Type:        formValues.Type,
		models.MessageColumns.Language:    formValues.Language,
	})
	if err != nil {
		errors.Add("form", "Failed to update message")
		return kit.Render(messages.MessageForm(formValues, formSettings, errors))
	}

	err = upsertMessageWebsites(kit.Request.Context(), messageId, formValues.Websites)
	if err != nil {
		errors.Add("form", "Failed to update message websites")
		return kit.Render(messages.MessageForm(formValues, formSettings, errors))
	}

	return kit.Redirect(200, "/messages")
}

func upsertMessageWebsites(ctx context.Context, messageId int64, websites []string) error {
	_, err := models.WebsitesMessages(
		models.WebsitesMessageWhere.MessageId.EQ(messageId),
	).DeleteAll(ctx, db.Query)
	if err != nil {
		return err
	}

	for _, websiteId := range websites {
		websiteIdInt, err := strconv.ParseInt(websiteId, 10, 64)
		if err != nil {
			return err
		}

		websiteMessage := &models.WebsitesMessage{
			WebsiteId: websiteIdInt,
			MessageId: messageId,
		}

		err = websiteMessage.Insert(ctx, db.Query, boil.Infer())
		if err != nil {
			return err
		}
	}

	return nil
}

func HandleMessageDelete(kit *kit.Kit) error {
	messageId, err := helpers.GetIdFromUrl(kit)
	if err != nil {
		return err
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

func parseMultiSelectFields(r *http.Request, data any) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %v", err)
	}

	val := reflect.ValueOf(data).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		formTag := field.Tag.Get("form")

		fieldVal := val.Field(i)
		if fieldVal.Kind() == reflect.Slice {
			formValues := r.Form[formTag]
			if formValues == nil {
				continue
			}
			if fieldVal.Type().Elem().Kind() == reflect.String {
				slice := reflect.MakeSlice(fieldVal.Type(), len(formValues), len(formValues))
				for j, val := range formValues {
					slice.Index(j).SetString(val)
				}
				fieldVal.Set(slice)
			} else {
				return fmt.Errorf("unsupported slice element kind %s", fieldVal.Type().Elem().Kind())
			}
		}
	}

	return nil
}
