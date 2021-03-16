package manage

import (
	"app/api"
	"app/datastore"

	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const TEMPLATE_DIR = "cmd/templates/manage/"

func Register() error {
	fs := http.FileServer(http.Dir("cmd/public"))

	http.Handle("/manage/js/", fs)
	http.Handle("/manage/css/", fs)

	mr := mux.NewRouter()
	h := NewHandler(mr)

	mr.HandleFunc("/manage/home", h.View).Methods("GET")
	//Page
	mr.HandleFunc("/manage/page/", h.ViewRootPage).Methods("GET")
	mr.HandleFunc("/manage/page/{key}", h.ViewPage)
	mr.HandleFunc("/manage/page/add/{key}", h.AddPage).Methods("GET")
	mr.HandleFunc("/manage/page/delete/{key}", h.DeletePage).Methods("GET")
	mr.HandleFunc("/manage/page/public/{key}", h.PublicPage).Methods("GET")
	mr.HandleFunc("/manage/page/private/{key}", h.PrivatePage).Methods("GET")
	mr.HandleFunc("/manage/page/tool/{key}", h.ToolPage).Methods("GET")
	mr.HandleFunc("/manage/page/tool/sequence", h.SequencePage).Methods("POST")
	mr.HandleFunc("/manage/page/tree/", h.TreePage).Methods("GET")

	//ページ表示
	mr.HandleFunc("/manage/page/view/{key}", h.PageHandler).Methods("GET")
	mr.HandleFunc("/manage/page/view/", h.TopHandler).Methods("GET")

	//File
	mr.HandleFunc("/manage/file/", h.ViewFile).Methods("GET")
	mr.HandleFunc("/manage/file/type/{type}", h.ViewFile).Methods("GET")
	mr.HandleFunc("/manage/file/add", h.AddFile).Methods("POST")
	mr.HandleFunc("/manage/file/delete", h.DeleteFile).Methods("POST")
	mr.HandleFunc("/manage/file/resize/{key}", h.ResizeFile).Methods("GET")
	mr.HandleFunc("/manage/file/resize/commit", h.ResizeCommitFile).Methods("POST")
	mr.HandleFunc("/manage/file/resize/view/{key}", h.ResizeFileView).Methods("GET")

	//Template
	mr.HandleFunc("/manage/template/", h.ViewTemplate).Methods("GET")
	mr.HandleFunc("/manage/template/add", h.AddTemplate).Methods("GET")
	mr.HandleFunc("/manage/template/edit/{key}", h.EditTemplate)
	mr.HandleFunc("/manage/template/delete/{key}", h.DeleteTemplate)
	mr.HandleFunc("/manage/template/reference/{key}", h.ReferenceTemplate)

	//table
	mr.HandleFunc("/manage/table/view", h.TableView)

	mr.HandleFunc("/manage/datastore/backup", h.Backup).Methods("POST")
	mr.HandleFunc("/manage/datastore/restore", h.Restore).Methods("POST")

	//Site
	mr.HandleFunc("/manage/site/", h.ViewSetting).Methods("GET")
	mr.HandleFunc("/manage/site/edit", h.EditSetting).Methods("POST")
	mr.HandleFunc("/manage/site/map", h.DownloadSitemap).Methods("GET")

	http.Handle("/manage/", h)

	return nil
}

type Handler struct {
	r *mux.Router
}

func NewHandler(r *mux.Router) Handler {
	return Handler{r}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Println("ServeHTTP:" + r.URL.String())
	//セッションの存在を確認
	u, err := GetSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", 301)
		return
	}

	if u == nil {
		log.Println("ユーザがいない")
		http.Redirect(w, r, "/login", 301)
		return
	}

	h.r.ServeHTTP(w, r)
}

func (h Handler) View(w http.ResponseWriter, r *http.Request) {
	h.parse(w, TEMPLATE_DIR+"top.tmpl", nil)
}

func (h Handler) TopHandler(w http.ResponseWriter, r *http.Request) {

	site, err := datastore.SelectSite(r, -1)
	if err != nil {
		if err == datastore.SiteNotFoundError {
			h.ViewSetting(w, r)
			return
		}
		h.errorPage(w, "Not Found", "Root page not found", 404)
		return
	}
	h.pageView(w, r, site.Root)
}

func (h Handler) PageHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]

	h.pageView(w, r, id)
}

func (h Handler) pageView(w http.ResponseWriter, r *http.Request, id string) {

	page := 1
	val := r.URL.Query()
	pageVal := val.Get("page")
	if pageVal != "" {
		p, err := strconv.Atoi(pageVal)
		if err == nil {
			page = p
		}
	}

	//管理用の書き出し
	err := datastore.WriteManageHTML(w, r, id, page)
	if err != nil {
		h.errorPage(w, "ERROR:Generate HTML", err.Error(), 500)
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
		log.Println(err)
		//h.errorPage(w, "Template Parse Error", err.Error(), 500)
		return
	}

	err = tmpl.Execute(w, obj)
	if err != nil {
		log.Println(err)
		//h.errorPage(w, "Template Execute Error", err.Error(), 500)
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
