package handlers

import (
	"messages/app/db"
	"messages/app/helpers"
	"messages/app/models"
	"messages/app/views/websites"
	"messages/plugins/auth"

	v "github.com/anthdm/superkit/validate"

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
	websiteId, err := helpers.GetIdFromUrl(kit)
	if err != nil {
		return err
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

	if err := helpers.VerifyAdminRole(kit.Auth().(auth.Auth)); err != nil {
		errors.Add("form", "You are not allowed to create websites")
		return kit.Render(websites.WebsiteForm(formValues, errors))
	}

	if errors, ok := v.Request(kit.Request, formValues, createWebsiteSchema); !ok {
		return kit.Render(websites.WebsiteForm(formValues, errors))
	}

	dbWebsite := models.Website{
		Name:    formValues.Name,
		URL:     formValues.Domain,
		Staging: formValues.Staging,
	}

	if err := dbWebsite.Insert(kit.Request.Context(), db.Query, boil.Infer()); err != nil {
		errors.Add("form", "Failed to create website")
		return kit.Render(websites.WebsiteForm(formValues, errors))
	}

	return kit.Redirect(200, "/websites")
}

func HandleWebsiteUpdate(kit *kit.Kit) error {
	errors := v.Errors{}
	formValues := getBaseWebsiteFormValues()
	var err error

	if formValues.ID, err = helpers.GetIdFromUrl(kit); err != nil {
		return err
	}

	if err := helpers.VerifyAdminRole(kit.Auth().(auth.Auth)); err != nil {
		errors.Add("form", "You are not allowed to update websites")
		return kit.Render(websites.WebsiteForm(formValues, errors))
	}

	if errors, ok := v.Request(kit.Request, formValues, createWebsiteSchema); !ok {
		return kit.Render(websites.WebsiteForm(formValues, errors))
	}

	if _, err := models.Websites(
		models.WebsiteWhere.ID.EQ(formValues.ID),
	).UpdateAll(kit.Request.Context(), db.Query, models.M{
		models.WebsiteColumns.Name:    formValues.Name,
		models.WebsiteColumns.URL:     formValues.Domain,
		models.WebsiteColumns.Staging: formValues.Staging,
	}); err != nil {
		errors.Add("form", "Failed to update website")
		return kit.Render(websites.WebsiteForm(formValues, errors))
	}

	return kit.Redirect(200, "/websites")
}

func HandleWebsiteDelete(kit *kit.Kit) error {
	errors := v.Errors{}
	confirmationModalProps := websites.GetDefaultConfirmationModalProps(kit.Request.Context())

	websiteId, err := helpers.GetIdFromUrl(kit)
	if err != nil {
		return err
	}

	if err := helpers.VerifyAdminRole(kit.Auth().(auth.Auth)); err != nil {
		errors.Add("form", "You are not allowed to delete websites")
		confirmationModalProps.Errors = errors
		return kit.Render(websites.ConfirmationModalContent(confirmationModalProps))
	}

	if _, err := models.Websites(
		models.WebsiteWhere.ID.EQ(websiteId),
	).DeleteAll(kit.Request.Context(), db.Query); err != nil {
		errors.Add("form", "Failed to delete website")
		confirmationModalProps.Errors = errors
		return kit.Render(websites.ConfirmationModalContent(confirmationModalProps))
	}

	return kit.Redirect(200, "/websites")
}

func getBaseWebsiteFormValues() *websites.WebsiteFormValues {
	return &websites.WebsiteFormValues{
		Name:   "",
		Domain: "",
	}
}
