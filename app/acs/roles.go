package acs

import (
	"context"
	"messages/plugins/auth"

	"github.com/invopop/ctxi18n/i18n"
)

func GetRolesList(ctx context.Context) map[string]string {
	return map[string]string{
		RoleUser:  i18n.T(ctx, "users.roles.values.user"),
		RoleAdmin: i18n.T(ctx, "users.roles.values.admin"),
	}
}

func GetRoleName(ctx context.Context, role string) string {
	roles := GetRolesList(ctx)
	return roles[role]
}

func IsValidRole(role string) bool {
	validRoles := []string{RoleUser, RoleAdmin}
	for _, r := range validRoles {
		if r == role {
			return true
		}
	}
	return false
}

func HasMinimumRole(auth auth.Auth, role string) bool {
	if auth.Role == RoleAdmin {
		return true
	}
	return auth.Role == role
}

type Role struct {
	Name string
}

const (
	RoleUser  string = "user"
	RoleAdmin string = "admin"
)
