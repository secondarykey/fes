package api

const (
	SiteTemplateName = "site_template"
	PageTemplateName = "page_template"
)

func ConvertTemplateType(data int) string {
	if data == 1 {
		return "Site"
	}
	return "Page"
}
