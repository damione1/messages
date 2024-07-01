package handlers

import (
	"messages/app/db"
	"messages/app/models"
	"messages/app/views/users"
	"messages/plugins/auth"
	"strconv"

	v "github.com/anthdm/superkit/validate"

	"github.com/go-chi/chi/v5"

	"github.com/anthdm/superkit/kit"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func HandleUsersList(kit *kit.Kit) error {
	data := &users.IndexPageData{
		FormValues: &users.UserFormValues{},
	}

	dbUsersList, err := models.Users().All(kit.Request.Context(), db.Query)
	if err != nil {
		return err
	}

	usersList := make([]*users.UserListItem, 0, len(dbUsersList))
	for _, user := range dbUsersList {
		usersList = append(usersList, &users.UserListItem{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Role:      user.Role,
		})
	}
	data.UsersList = usersList

	return kit.Render(users.Index(data))
}

var createUserSchema = v.Schema{
	"email": v.Rules(v.Required),
}

func HandleUserInvite(kit *kit.Kit) error {
	auth := kit.Auth().(auth.Auth)
	formValues := &users.UserFormValues{}
	errors, ok := v.Request(kit.Request, formValues, createUserSchema)
	if !ok {
		return kit.Render(users.UserForm(formValues, errors))
	}

	invitation := models.Invitation{
		Email:     formValues.Email,
		InvitedBy: int64(auth.UserID),
	}

	err := invitation.Insert(kit.Request.Context(), db.Query, boil.Infer())
	if err != nil {
		return err
	}

	return kit.Redirect(200, "/users")
}

func HandleUserDelete(kit *kit.Kit) error {
	auth := kit.Auth().(auth.Auth)
	errors := v.Errors{}

	userId, err := strconv.ParseInt(chi.URLParam(kit.Request, "id"), 10, 64)
	if err != nil {
		errors.Add("form", "Not found")
		return kit.Render(users.DeleteConfirmationModal(0, errors))
	}

	if auth.Role != "admin" {
		errors.Add("form", "You do not have permission to delete users")
		return kit.Render(users.DeleteConfirmationModal(userId, errors))
	}

	user, err := models.Users(
		models.UserWhere.ID.EQ(userId),
	).One(kit.Request.Context(), db.Query)
	if err != nil {
		errors.Add("form", "Internal error")
		return kit.Render(users.DeleteConfirmationModal(userId, errors))
	}

	_, err = user.Delete(kit.Request.Context(), db.Query)
	if err != nil {
		errors.Add("form", "Internal error")
		return kit.Render(users.DeleteConfirmationModal(userId, errors))
	}

	return kit.Redirect(200, "/users")
}
