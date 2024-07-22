package handlers

import (
	"errors"
	"messages/app/db"
	"messages/app/helpers"
	"messages/app/models"
	"messages/app/views/users"
	"messages/plugins/auth"
	"strconv"

	v "github.com/anthdm/superkit/validate"

	"github.com/go-chi/chi/v5"

	"github.com/anthdm/superkit/kit"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func HandleUsersList(kit *kit.Kit) error {
	data := &users.IndexPageData{
		FormValues:     &users.InvitationFormValues{},
		InvitationList: make([]*users.InvitationListItem, 0),
	}

	dbUsersList, err := models.Users().All(kit.Request.Context(), db.Query)
	if err != nil {
		return helpers.RenderNoticeError(kit, err)
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

	dbInvitationsList, err := models.Invitations(
		qm.Load(models.InvitationRels.InvitedByUser),
	).All(kit.Request.Context(), db.Query)
	if err != nil {
		return helpers.RenderNoticeError(kit, err)
	}

	invitationsList := make([]*users.InvitationListItem, 0, len(dbInvitationsList))
	for _, dbInvitation := range dbInvitationsList {
		invitation := &users.InvitationListItem{
			ID:    dbInvitation.ID,
			Email: dbInvitation.Email,
		}
		if dbInvitation.R.InvitedByUser != nil {
			invitation.InvitedBy = dbInvitation.R.InvitedByUser.FirstName + " " + dbInvitation.R.InvitedByUser.LastName
		}
		invitationsList = append(invitationsList, invitation)
	}

	data.InvitationList = invitationsList

	return kit.Render(users.Index(data))
}

var createUserSchema = v.Schema{
	"email": v.Rules(v.Required, v.Email),
}

func HandleInvitationCreate(kit *kit.Kit) error {
	auth := kit.Auth().(auth.Auth)
	formValues := &users.InvitationFormValues{}
	errors := v.Errors{}

	if auth.Role != "admin" {
		errors.Add("form", "You do not have permission to invite users")
		return kit.Render(users.InvitationForm(formValues, errors))
	}

	errors, ok := v.Request(kit.Request, formValues, createUserSchema)
	if !ok {
		return kit.Render(users.InvitationForm(formValues, errors))
	}

	ok, err := models.Users(
		models.UserWhere.Email.EQ(formValues.Email),
	).Exists(kit.Request.Context(), db.Query)
	if err != nil {
		errors.Add("form", "Internal error")
		return kit.Render(users.InvitationForm(formValues, errors))
	}
	if ok {
		errors.Add("form", "User already exists")
		return kit.Render(users.InvitationForm(formValues, errors))
	}

	ok, err = models.Invitations(
		models.InvitationWhere.Email.EQ(formValues.Email),
	).Exists(kit.Request.Context(), db.Query)
	if err != nil {
		errors.Add("form", "Internal error")
		return kit.Render(users.InvitationForm(formValues, errors))
	}
	if ok {
		errors.Add("form", "User already invited")
		return kit.Render(users.InvitationForm(formValues, errors))
	}

	invitation := models.Invitation{
		Email:     formValues.Email,
		InvitedBy: int64(auth.UserID),
	}

	err = invitation.Insert(kit.Request.Context(), db.Query, boil.Infer())
	if err != nil {
		return err
	}

	return kit.Redirect(200, "/users")
}

func HandleUserDelete(kit *kit.Kit) error {
	auth := kit.Auth().(auth.Auth)

	userId, err := strconv.ParseInt(chi.URLParam(kit.Request, "id"), 10, 64)
	if err != nil {
		return helpers.RenderNoticeError(kit, errors.New("Not found"))
	}

	if auth.Role != "admin" {
		return helpers.RenderNoticeError(kit, errors.New("You do not have permission to delete users"))
	}

	if int64(auth.UserID) == userId {
		return helpers.RenderNoticeError(kit, errors.New("You cannot delete yourself"))
	}

	user, err := models.Users(
		models.UserWhere.ID.EQ(userId),
	).One(kit.Request.Context(), db.Query)
	if err != nil {
		return helpers.RenderNoticeError(kit, err)
	}

	_, err = user.Delete(kit.Request.Context(), db.Query)
	if err != nil {
		return helpers.RenderNoticeError(kit, err)
	}

	return kit.Redirect(200, "/users")
}

func HandleInvitationDelete(kit *kit.Kit) error {
	auth := kit.Auth().(auth.Auth)

	if auth.Role != "admin" {
		return helpers.RenderNoticeError(kit, errors.New("You do not have permission to delete invitations"))
	}

	invitationId, err := strconv.ParseInt(chi.URLParam(kit.Request, "id"), 10, 64)
	if err != nil {
		return helpers.RenderNoticeError(kit, errors.New("Not found"))
	}

	invitation, err := models.Invitations(
		models.InvitationWhere.ID.EQ(invitationId),
	).One(kit.Request.Context(), db.Query)
	if err != nil {
		return helpers.RenderNoticeError(kit, err)
	}

	_, err = invitation.Delete(kit.Request.Context(), db.Query)
	if err != nil {
		return helpers.RenderNoticeError(kit, err)
	}

	return kit.Redirect(200, "/users")
}
