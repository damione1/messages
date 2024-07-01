package helpers

import (
	"regexp"

	"github.com/anthdm/superkit/validate"
)

func IsValidDomain(domain string) bool {
	// Regular expression to validate domain names
	var re = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`)
	return re.MatchString(domain)
}

var ValidDomain = validate.RuleSet{
	Name: "domain",
	MessageFunc: func(set validate.RuleSet) string {
		return "must be a valid domain (e.g. example.com)"
	},
	ValidateFunc: func(rule validate.RuleSet) bool {
		str, ok := rule.FieldValue.(string)
		if !ok {
			return false
		}
		return IsValidDomain(str)
	},
}
