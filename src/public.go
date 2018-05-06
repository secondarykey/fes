package src

import (
	"datastore"
	"html/template"
	"net/http"

	"api"
	"github.com/gorilla/mux"
	"log"
)

type Public struct {
	r *http.Request
}

func (p Public) topHandler(w http.ResponseWriter, r *http.Request) {
	top, err := datastore.SelectRootPage(r)
	if err != nil {
		p.errorPage(w, err.Error(), "Datastore Select Page [main]", 500)
		return
	}

	if top == nil {
		p.errorPage(w, err.Error(), "Not Found[main]", 404)
		return
	}
	p.pageParse(w, r, top)
}

func (p Public) pageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]
	page, err := datastore.SelectPage(r, id)
	if err != nil {
		p.errorPage(w, err.Error(), "Datastore Select Page ["+id+"]", 500)
		return
	}

	if page == nil {
		p.errorPage(w, "Not Found", "Not found page["+id+"]", 404)
		return
	}

	p.pageParse(w, r, page)
}

func (pub Public) pageParse(w http.ResponseWriter, r *http.Request, page *datastore.Page) {

	var err error

	id := page.Key.StringID()
	pub.r = r
	site := datastore.GetSite(r)

	//テンプレートを取得
	siteTmp, err := datastore.SelectTemplateData(r, page.SiteTemplate)
	if err != nil {
		pub.errorPage(w, err.Error(), "Datastore:Select Site Template Error", 500)
		return
	}
	pageTmp, err := datastore.SelectTemplateData(r, page.PageTemplate)
	if err != nil {
		pub.errorPage(w, err.Error(), "Datastore:Select Page Template Error", 500)
		return
	}

	pData, err := datastore.SelectPageData(r, id)
	if err != nil {
		pub.errorPage(w, err.Error(), "Datastore:Select Page Data Error", 500)
		return
	}
	children, err := datastore.SelectChildPages(r, id)
	if err != nil {
		pub.errorPage(w, "Datastore:Select Children page Error", err.Error(), 500)
		return
	}

	siteTmpData := string(siteTmp.Content)
	pageTmpData := string(pageTmp.Content)
	siteTmpData = "{{define \"" + api.SITE_TEMPLATE + "\"}}" + "\n" + siteTmpData + "\n" + "{{end}}"
	pageTmpData = "{{define \"" + api.PAGE_TEMPLATE + "\"}}" + "\n" + pageTmpData + "\n" + "{{end}}"

	funcMap := template.FuncMap{
		"html":        api.ConvertHTML,
		"plane":       api.ConvertString,
		"convertDate": api.ConvertDate,
		"list":        pub.list,
	}

	//適用する
	tmpl, err := template.New(api.SITE_TEMPLATE).Funcs(funcMap).Parse(siteTmpData)
	if err != nil {
		pub.errorPage(w, err.Error(), "Template:Parse Site Template Error", 500)
		return
	}
	tmpl, err = tmpl.Parse(pageTmpData)
	if err != nil {
		pub.errorPage(w, err.Error(), "Template:Parse Page Template Error", 500)
		return
	}

	dto := struct {
		Site     *datastore.Site
		Page     *datastore.Page
		PageData *datastore.PageData
		Children []datastore.Page
	}{site, page, pData, children}

	err = tmpl.Execute(w, dto)
	if err != nil {
		pub.errorPage(w, err.Error(), "Template:Exevute Page Data Error", 500)
		return
	}
}

func (p Public) list(id string) []datastore.Page {
	pages, err := datastore.SelectChildPages(p.r, id)
	if err != nil {
		return make([]datastore.Page, 0)
	}
	return pages
}

func (p Public) fileHandler(w http.ResponseWriter, r *http.Request) {

	//ファイルを検索
	vars := mux.Vars(r)
	id := vars["key"]

	//表示
	fileData, err := datastore.SelectFileData(r, id)
	if err != nil {
		p.errorPage(w, err.Error(), "Datastore:FileData Search Error", 500)
		return
	}

	if fileData == nil {
		p.errorPage(w, err.Error(), "Datastore:Not Found FileData Error", 404)
		return
	}

	w.Header().Set("Content-Type", fileData.Mime)
	_, err = w.Write(fileData.Content)
	if err != nil {
		p.errorPage(w, "Writing FileData Error", err.Error(), 500)
		return
	}
}

func (p Public) errorPage(w http.ResponseWriter, t string, msg string, num int) {

	dto := struct {
		Title   string
		Message string
		No      int
	}{t, msg, num}

	w.WriteHeader(num)

	tmpl, err := template.ParseFiles("templates/error.tmpl")
	if err != nil {
		log.Println("Error Page Parse Error")
		log.Println(err)
		return
	}
	err = tmpl.Execute(w, dto)
	if err != nil {
		log.Println("Error Page Execute Error")
		log.Println(err)
		return
	}
}
