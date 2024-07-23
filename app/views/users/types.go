package users

import v "github.com/anthdm/superkit/validate"

type UserListItem struct {
	ID        int64
	FirstName string
	LastName  string
	Email     string
	Role      string
}

type InvitationListItem struct {
	ID        int64
	Email     string
	InvitedBy string
}

type IndexPageData struct {
	UsersList      []*UserListItem
	InvitationList []*InvitationListItem
	FormValues     *InvitationFormValues
	FormErrors     v.Errors
}

type InvitationFormValues struct {
	Email string `form:"email"`
}

type UpdateUserRoleFormValues struct {
	Role string `form:"role"`
}
