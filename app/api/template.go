package api

const (
	SiteTemplateName = "site_template"
	PageTemplateName = "page_template"
)

const (
	FileTypeData      = 1
	FileTypePageImage = 2
)

func ConvertTemplateType(data int) string {
	if data == 1 {
		return "Site"
	}
	return "Page"
}
