package auth

import (
	"fmt"
	"messages/app/db"
	"messages/app/models"

	"github.com/anthdm/superkit/kit"
	v "github.com/anthdm/superkit/validate"
)

var profileSchema = v.Schema{
	"firstName": v.Rules(v.Min(3), v.Max(50)),
	"lastName":  v.Rules(v.Min(3), v.Max(50)),
}

type ProfileFormValues struct {
	ID        int    `form:"id"`
	FirstName string `form:"firstName"`
	LastName  string `form:"lastName"`
	Email     string `form:"email"`
	Success   string
	Role      string
}

func HandleProfileShow(kit *kit.Kit) error {
	auth := kit.Auth().(Auth)

	user, err := models.Users(
		models.UserWhere.ID.EQ(int64(auth.UserID)),
	).One(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}

	formValues := ProfileFormValues{
		ID:        int(user.ID),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      auth.Role,
	}

	return kit.Render(ProfileShow(formValues))
}

func HandleProfileUpdate(kit *kit.Kit) error {
	var values ProfileFormValues
	errors, ok := v.Request(kit.Request, &values, profileSchema)
	if !ok {
		return kit.Render(ProfileForm(values, errors))
	}

	auth := kit.Auth().(Auth)
	if auth.UserID != values.ID {
		return fmt.Errorf("unauthorized request for profile %d", values.ID)
	}

	_, err := models.Users(
		models.UserWhere.ID.EQ(int64(auth.UserID)),
	).UpdateAll(kit.Request.Context(), db.Query, models.M{
		models.UserColumns.FirstName: values.FirstName,
		models.UserColumns.LastName:  values.LastName,
		models.UserColumns.Email:     values.Email,
	})
	if err != nil {
		return err
	}

	values.Success = "Profile successfully updated!"

	return kit.Render(ProfileForm(values, v.Errors{}))
}
