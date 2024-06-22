package handlers

import (
	"errors"
	"messages/app/db"
	"messages/app/models"

	websitesView "messages/app/views/websites"

	"github.com/anthdm/superkit/kit"
)

func HandleSites(kit *kit.Kit) error {
	sitesList, err := models.Websites().All(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}
	return kit.Render(websitesView.Index(sitesList))
}

func HandleSite(kit *kit.Kit) error {
	return errors.New("Not implemented")
}
