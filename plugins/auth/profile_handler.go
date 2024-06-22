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
	Email     string
	Success   string
}

func HandleProfileShow(kit *kit.Kit) error {
	auth := kit.Auth().(Auth)

	// var user User
	// err := db.Query.NewSelect().
	// 	Model(&user).
	// 	Where("id = ?", auth.UserID).
	// 	Scan(kit.Request.Context())
	// if err != nil {
	// 	return err
	// }
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
	// _, err := db.Query.NewUpdate().
	// 	Model((*User)(nil)).
	// 	Set("first_name = ?", values.FirstName).
	// 	Set("last_name = ?", values.LastName).
	// 	Where("id = ?", auth.UserID).
	// 	Exec(kit.Request.Context())
	// if err != nil {
	// 	return err
	// }
	_, err := models.Users(
		models.UserWhere.ID.EQ(int64(auth.UserID)),
	).UpdateAll(kit.Request.Context(), db.Query, models.M{
		models.UserColumns.FirstName: values.FirstName,
		models.UserColumns.LastName:  values.LastName,
	})
	if err != nil {
		return err
	}

	values.Success = "Profile successfully updated!"
	values.Email = auth.Email

	return kit.Render(ProfileForm(values, v.Errors{}))
}
