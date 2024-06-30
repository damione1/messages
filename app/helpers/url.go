package helpers

import "regexp"

func IsValidDomain(domain string) bool {
	// Regular expression to validate domain names
	var re = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`)
	return re.MatchString(domain)
}
