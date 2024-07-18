package locales

import "embed"

//go:embed *.yaml
var LocalesFs embed.FS

var LanguageList = []string{
	"fr",
	"en",
}
