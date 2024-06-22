package handlers

import (
	"messages/app/views/index"

	"github.com/anthdm/superkit/kit"
)

func HandleLandingIndex(kit *kit.Kit) error {
	return kit.Render(index.Index())
}
