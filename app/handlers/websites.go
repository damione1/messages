package handlers

import (
	"errors"
	"messages/app/db"
	"messages/app/models"
	"messages/app/views/websites"
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
			ID:   website.ID,
			Name: website.WebsiteName,
			URL:  website.WebsiteUrl,
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

	data.FormValues.Name = dbWebsite.WebsiteName
	data.FormValues.URL = dbWebsite.WebsiteUrl
	data.FormValues.ID = dbWebsite.ID

	return kit.Render(websites.PageWebsiteEdit(data))
}

var createWebsiteSchema = v.Schema{
	"name": v.Rules(v.Required),
	"url":  v.Rules(v.Required),
}

func HandleWebsiteCreate(kit *kit.Kit) error {
	formValues := getBaseWebsiteFormValues()
	errors, ok := v.Request(kit.Request, &formValues, createWebsiteSchema)
	if !ok {
		return kit.Render(websites.WebsiteForm(formValues, errors))
	}

	website := models.Website{
		WebsiteName: formValues.Name,
		WebsiteUrl:  formValues.URL,
	}

	err := website.Insert(kit.Request.Context(), db.Query, boil.Infer())
	if err != nil {
		return err
	}

	return kit.Redirect(200, "/websites")
}

func HandleWebsiteUpdate(kit *kit.Kit) error {
	paramId := chi.URLParam(kit.Request, "id")

	websiteId, err := strconv.ParseInt(paramId, 10, 64)
	if err != nil {
		return errors.New("Not found")
	}

	formValues := getBaseWebsiteFormValues()
	formValues.ID = websiteId

	errors, ok := v.Request(kit.Request, &formValues, createWebsiteSchema)
	if !ok {
		return kit.Render(websites.WebsiteForm(formValues, errors))
	}

	website, err := models.Websites(
		models.WebsiteWhere.ID.EQ(formValues.ID),
	).One(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}

	website.WebsiteName = formValues.Name
	website.WebsiteUrl = formValues.URL

	_, err = website.Update(kit.Request.Context(), db.Query, boil.Infer())
	if err != nil {
		return err
	}

	return kit.Redirect(200, "/websites")
}

func HandleWebsiteDelete(kit *kit.Kit) error {
	paramId := chi.URLParam(kit.Request, "id")

	websiteId, err := strconv.ParseInt(paramId, 10, 64)
	if err != nil {
		return errors.New("Not found")
	}

	website, err := models.Websites(
		models.WebsiteWhere.ID.EQ(websiteId),
	).One(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}

	_, err = website.Delete(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}

	return kit.Redirect(200, "/websites")
}

func getBaseWebsiteFormValues() websites.WebsiteFormValues {
	return websites.WebsiteFormValues{
		Name: "",
		URL:  "",
	}
}
