package manage

import (
	"api"
	"html/template"
	"net/http"
)

const TEMPLATE_DIR = "./templates/manage/"

type Handler struct{}

func (h Handler) View(w http.ResponseWriter, r *http.Request) {

	//特にないかな？

	h.parse(w, TEMPLATE_DIR+"top.tmpl", nil)
}

func (h Handler) parse(w http.ResponseWriter, tName string, obj interface{}) {

	funcMap := template.FuncMap{
		"plane":               api.ConvertString,
		"html":                api.ConvertHTML,
		"convertDate":         api.ConvertDate,
		"convertSize":         api.ConvertSize,
		"convertTemplateType": convertTemplateType,
	}
	tmpl, err := template.New(api.SITE_TEMPLATE).Funcs(funcMap).ParseFiles(TEMPLATE_DIR+"layout.tmpl", tName)
	if err != nil {
		h.errorPage(w, "Template Parse Error", err.Error(), 500)
		return
	}

	err = tmpl.Execute(w, obj)
	if err != nil {
		h.errorPage(w, "Template Execute Error", err.Error(), 500)
		return
	}
}

func (h Handler) errorPage(w http.ResponseWriter, t string, e string, num int) {
	dto := struct {
		Title       string
		Description string
		Number      int
	}{t, e, num}

	h.parse(w, TEMPLATE_DIR+"error.tmpl", dto)
	w.WriteHeader(num)
}

func convertTemplateType(data int) string {
	if data == 1 {
		return "Site"
	}
	return "Page"
}

func POST(r *http.Request) bool {
	if r.Method == "POST" {
		return true
	}
	return false
}

func GET(r *http.Request) bool {
	if r.Method == "GET" {
		return true
	}
	return false
}
