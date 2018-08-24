package manage

import (
	"api"
	"datastore"

	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"strconv"
)

const TEMPLATE_DIR = "./templates/manage/"

type Handler struct{}

func (h Handler) View(w http.ResponseWriter, r *http.Request) {
	h.parse(w, TEMPLATE_DIR+"top.tmpl", nil)
}

func (h Handler) TopHandler(w http.ResponseWriter, r *http.Request) {
	site,err := datastore.SelectSite(r)
	if err != nil {
		if err == datastore.SiteNotFoundError {
			h.ViewSetting(w,r)
			return
		}
		h.errorPage(w,"Not Found","Root page not found",404)
		return
	}
	h.pageView(w,r,site.Root)
}

func (h Handler) PageHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]

	h.pageView(w,r,id)
}

func (h Handler) pageView(w http.ResponseWriter, r *http.Request,id string) {

	page := 1
	val := r.URL.Query()
	pageVal := val.Get("page")
	if pageVal != "" {
		p,err := strconv.Atoi(pageVal)
		if err == nil {
			page = p
		}
	}

	//管理用の書き出し
	err := datastore.WriteManageHTML(w,r,id,page)
	if err != nil {
		h.errorPage(w,"ERROR:Generate HTML",err.Error(),500)
		return
	}
}

func (h Handler) parse(w http.ResponseWriter, tName string, obj interface{}) {

	funcMap := template.FuncMap{
		"plane":               api.ConvertString,
		"html":                api.ConvertHTML,
		"convertDate":         api.ConvertDate,
		"convertSize":         api.ConvertSize,
		"convertTemplateType": convertTemplateType,
	}
	tmpl, err := template.New(api.SiteTemplateName).Funcs(funcMap).ParseFiles(TEMPLATE_DIR+"layout.tmpl", tName)
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
