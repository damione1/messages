package helpers

import (
	"errors"
	"messages/plugins/auth"
	"strconv"

	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
)

func GetIdFromUrl(kit *kit.Kit) (int64, error) {
	paramId := chi.URLParam(kit.Request, "id")
	websiteId, err := strconv.ParseInt(paramId, 10, 64)
	if err != nil {
		return 0, errors.New("Invalid website ID")
	}
	return websiteId, nil
}

func VerifyAdminRole(auth auth.Auth) error {
	if auth.Role != "admin" {
		return errors.New("You are not allowed to perform this action")
	}
	return nil
}
