package handlers

import (
	"messages/app/views/index"

	"github.com/anthdm/superkit/kit"
)

func HandleApi(kit *kit.Kit) error {
	return kit.Render(index.Index())
}
