package websites

import v "github.com/anthdm/superkit/validate"

type IndexPageData struct {
	WebsitesList []*WebsiteListItem
	FormValues   *WebsiteFormValues
	FormErrors   v.Errors
}

type PageWebsiteEditData struct {
	FormValues *WebsiteFormValues
	FormErrors v.Errors
}

type WebsiteListItem struct {
	ID      int64
	Name    string
	Domain  string
	Staging bool
}

type WebsiteFormValues struct {
	ID      int64  `form:"id"`
	Name    string `form:"name"`
	Domain  string `form:"domain"`
	Staging bool   `form:"staging"`
}

type ConfirmationModalProps struct {
	Title   string
	Message string
	Action  string
	Errors  v.Errors
}
