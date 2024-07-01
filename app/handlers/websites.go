package handlers

import (
	"errors"
	"messages/app/db"
	"messages/app/helpers"
	"messages/app/models"
	"messages/app/views/websites"
	"messages/plugins/auth"
	"strconv"

	v "github.com/anthdm/superkit/validate"
	"github.com/go-chi/chi/v5"

	"github.com/anthdm/superkit/kit"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func HandleWebsitesList(kit *kit.Kit) error {
	data := &websites.IndexPageData{
		FormValues: getBaseWebsiteFormValues(),
	}

	dbWebsitesList, err := models.Websites().All(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}

	websitesList := make([]*websites.WebsiteListItem, 0, len(dbWebsitesList))
	for _, website := range dbWebsitesList {
		websitesList = append(websitesList, &websites.WebsiteListItem{
			ID:      website.ID,
			Name:    website.Name,
			Domain:  website.URL,
			Staging: website.Staging,
		})
	}
	data.WebsitesList = websitesList

	return kit.Render(websites.Index(data))
}

func HandleWebsiteGet(kit *kit.Kit) error {
	paramId := chi.URLParam(kit.Request, "id")

	websiteId, err := strconv.ParseInt(paramId, 10, 64)
	if err != nil {
		return errors.New("Not found")
	}
	// // Get the website from the database
	dbWebsite, err := models.FindWebsite(kit.Request.Context(), db.Query, websiteId)
	if err != nil {
		return err
	}

	data := &websites.PageWebsiteEditData{
		FormValues: getBaseWebsiteFormValues(),
		FormErrors: v.Errors{},
	}

	data.FormValues.Name = dbWebsite.Name
	data.FormValues.Domain = dbWebsite.URL
	data.FormValues.ID = dbWebsite.ID
	data.FormValues.Staging = dbWebsite.Staging

	return kit.Render(websites.PageWebsiteEdit(data))
}

var createWebsiteSchema = v.Schema{
	"name":    v.Rules(v.Required),
	"domain":  v.Rules(v.Required, helpers.ValidDomain),
	"staging": v.Rules(),
}

func HandleWebsiteCreate(kit *kit.Kit) error {
	formValues := getBaseWebsiteFormValues()
	errors := v.Errors{}

	auth := kit.Auth().(auth.Auth)
	if auth.Role != "admin" {
		errors.Add("form", "You are not allowed to create websites")
		return kit.Render(websites.WebsiteForm(formValues, errors))
	}

	errors, ok := v.Request(kit.Request, &formValues, createWebsiteSchema)
	if !ok {
		return kit.Render(websites.WebsiteForm(formValues, errors))
	}

	dbWebsite := models.Website{
		Name:    formValues.Name,
		URL:     formValues.Domain,
		Staging: formValues.Staging,
	}

	err := dbWebsite.Insert(kit.Request.Context(), db.Query, boil.Infer())
	if err != nil {
		errors.Add("form", "Failed to create website")
		return kit.Render(websites.WebsiteForm(formValues, errors))
	}

	return kit.Redirect(200, "/websites")
}

func HandleWebsiteUpdate(kit *kit.Kit) error {
	formValues := getBaseWebsiteFormValues()
	paramId := chi.URLParam(kit.Request, "id")
	errors := v.Errors{}
	auth := kit.Auth().(auth.Auth)
	if auth.Role != "admin" {
		errors.Add("form", "You are not allowed to create websites")
		return kit.Render(websites.WebsiteForm(formValues, errors))
	}

	websiteId, err := strconv.ParseInt(paramId, 10, 64)
	if err != nil {
		errors.Add("form", "Invalid website ID")
		return kit.Render(websites.WebsiteForm(formValues, errors))
	}
	formValues.ID = websiteId

	errors, ok := v.Request(kit.Request, &formValues, createWebsiteSchema)
	if !ok {
		return kit.Render(websites.WebsiteForm(formValues, errors))
	}

	dbWebsite, err := models.Websites(
		models.WebsiteWhere.ID.EQ(formValues.ID),
	).One(kit.Request.Context(), db.Query)
	if err != nil {
		errors.Add("form", "Website not found")
		return kit.Render(websites.WebsiteForm(formValues, errors))
	}

	dbWebsite.Name = formValues.Name
	dbWebsite.URL = formValues.Domain
	dbWebsite.Staging = formValues.Staging

	_, err = dbWebsite.Update(kit.Request.Context(), db.Query, boil.Infer())
	if err != nil {
		errors.Add("form", "Failed to update website")
		return kit.Render(websites.WebsiteForm(formValues, errors))
	}

	return kit.Redirect(200, "/websites")
}

func HandleWebsiteDelete(kit *kit.Kit) error {
	paramId := chi.URLParam(kit.Request, "id")
	errors := v.Errors{}
	auth := kit.Auth().(auth.Auth)
	confirmationModalProps := websites.GetDefaultConfirmationModalProps()

	if auth.Role != "admin" {
		errors.Add("form", "You are not allowed to delete websites")
		confirmationModalProps.Errors = errors
		return kit.Render(websites.ConfirmationModalContent(confirmationModalProps))
	}

	websiteId, err := strconv.ParseInt(paramId, 10, 64)
	if err != nil {
		errors.Add("form", "Invalid website ID")
		confirmationModalProps.Errors = errors
		return kit.Render(websites.ConfirmationModalContent(confirmationModalProps))
	}

	website, err := models.Websites(
		models.WebsiteWhere.ID.EQ(websiteId),
	).One(kit.Request.Context(), db.Query)
	if err != nil {
		errors.Add("form", "Website not found")
		confirmationModalProps.Errors = errors
		return kit.Render(websites.ConfirmationModalContent(confirmationModalProps))
	}

	_, err = website.Delete(kit.Request.Context(), db.Query)
	if err != nil {
		errors.Add("form", "Failed to delete website")
		confirmationModalProps.Errors = errors
		return kit.Render(websites.ConfirmationModalContent(confirmationModalProps))
	}

	return kit.Redirect(200, "/websites")
}

func getBaseWebsiteFormValues() websites.WebsiteFormValues {
	return websites.WebsiteFormValues{
		Name:   "",
		Domain: "",
	}
}
