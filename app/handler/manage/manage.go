package manage

import (
	"app/datastore"
	. "app/handler/internal"
	"app/logic"

	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
)

func Register() error {

	err := RegisterManageStatic()
	if err != nil {
		return xerrors.Errorf("error: %w", err)
	}

	r := mux.NewRouter()
	s := r.PathPrefix("/manage").Subrouter()

	s.HandleFunc("/favico.ico", viewRootPageHandler).Methods("GET")

	//Page
	s.HandleFunc("/page/", viewRootPageHandler).Methods("GET")
	s.HandleFunc("/page/{key}", viewPageHandler)
	s.HandleFunc("/page/add/{key}", addPageHandler).Methods("GET")
	s.HandleFunc("/page/delete/{key}", deletePageHandler).Methods("GET")

	//Tool
	s.HandleFunc("/page/children/{key}", childrenPageHandler).Methods("GET")
	s.HandleFunc("/page/update/sequence", changeSequencePageHandler).Methods("POST")
	s.HandleFunc("/page/update/move", movePageHandler).Methods("POST")
	s.HandleFunc("/page/template/page/{key}", referencePageTemplateHandler).Methods("GET")
	s.HandleFunc("/page/template/site/{key}", referenceSiteTemplateHandler).Methods("GET")
	s.HandleFunc("/page/tree/", treePageHandler).Methods("GET")

	//HTML
	s.HandleFunc("/html/publish/{key}", changePublishPageHandler).Methods("GET")
	s.HandleFunc("/html/private/{key}", changePrivatePageHandler).Methods("GET")
	s.HandleFunc("/html/update", updateHTMLHandler).Methods("POST")

	//ページ表示
	s.HandleFunc("/page/view/{key}", privatePageHandler).Methods("GET")
	s.HandleFunc("/page/view/", privateHandler).Methods("GET")

	//File
	s.HandleFunc("/file/", viewFileHandler).Methods("GET")
	s.HandleFunc("/file/type/{type}", viewFileHandler).Methods("GET")
	s.HandleFunc("/file/add", addFileHandler).Methods("POST")
	s.HandleFunc("/file/favicon", faviconUploadHandler).Methods("POST")
	s.HandleFunc("/file/delete", deleteFileHandler).Methods("POST")
	s.HandleFunc("/file/resize/{key}", resizeFileHandler).Methods("GET")
	s.HandleFunc("/file/resize/commit", resizeCommitFileHandler).Methods("POST")
	s.HandleFunc("/file/resize/view/{key}", resizeFileViewHandler).Methods("GET")
	s.HandleFunc("/file/view/{key}", fileViewHandler).Methods("GET")

	//Template
	s.HandleFunc("/template/", viewTemplateHandler).Methods("GET")
	s.HandleFunc("/template/type/{type}", viewTemplateHandler).Methods("GET")
	s.HandleFunc("/template/add", addTemplateHandler).Methods("GET")
	s.HandleFunc("/template/edit/{key}", editTemplateHandler)
	s.HandleFunc("/template/delete/{key}", deleteTemplateHandler)
	s.HandleFunc("/template/reference/{type}/{key}", referenceTemplateHandler)

	s.HandleFunc("/variable/", viewVariableHandler).Methods("GET")
	s.HandleFunc("/variable/add", addVariableHandler).Methods("GET")
	s.HandleFunc("/variable/edit", editVariableHandler)
	s.HandleFunc("/variable/upload", uploadVariableHandler).Methods("POST")
	s.HandleFunc("/variable/edit/{key}", editVariableHandler).Methods("GET")
	s.HandleFunc("/variable/delete/{key}", deleteVariableHandler)

	//table
	s.HandleFunc("/table/view", viewTableHandler)

	s.HandleFunc("/datastore/backup", backupHandler).Methods("POST")
	s.HandleFunc("/datastore/restore", restoreHandler).Methods("POST")

	//Site
	s.HandleFunc("/site/", viewSiteHandler).Methods("GET")
	s.HandleFunc("/site/edit", editSiteHandler).Methods("POST")
	s.HandleFunc("/site/map", downloadSitemapHandler).Methods("GET")

	s.HandleFunc("/system/gc", gc).Methods("GET")

	s.HandleFunc("/", indexHandler).Methods("GET")

	h := NewHandler(s)
	http.Handle("/manage/", h)

	return nil
}

type ManageHandler struct {
	r *mux.Router
}

func NewHandler(r *mux.Router) ManageHandler {
	return ManageHandler{r}
}

func (h ManageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//セッションの存在を確認
	u, err := GetSession(r)
	if err != nil {
		log.Printf("%+v", err)
		http.Redirect(w, r, "/login", 302)
		return
	}

	if u == nil {
		log.Println("ユーザがいない")
		http.Redirect(w, r, "/login", 302)
		return
	}

	h.r.ServeHTTP(w, r)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	viewManage(w, "top.tmpl", nil)
}

func privateHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	dao := datastore.NewDao()

	site, err := dao.SelectSite(ctx, -1)
	if err != nil {
		if err == datastore.SiteNotFoundError {
			//TODO redirect???
			viewSiteHandler(w, r)
			return
		}
		errorPage(w, "Not Found", fmt.Errorf("Root page not found"), 404)
		return
	}
	pageView(w, r, site.Root)
}

func privatePageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]
	pageView(w, r, id)
}

func pageView(w http.ResponseWriter, r *http.Request, id string) {

	page := 1
	val := r.URL.Query()
	pageVal := val.Get("page")
	if pageVal != "" {
		p, err := strconv.Atoi(pageVal)
		if err == nil {
			page = p
		}
	}

	logic.ClearTemplateCache()

	//管理用の書き出し
	err := logic.WriteManageHTML(w, r, id, page, nil)
	if err != nil {
		errorPage(w, "ERROR:Generate HTML", err, 500)
		return
	}
}

func viewManage(w http.ResponseWriter, tName string, obj interface{}) {
	err := ViewManage(w, obj, tName)
	if err != nil {
		log.Printf("viewManage() error:\n%+v\n", err)
	}
	return
}

func gc(w http.ResponseWriter, r *http.Request) {

	err := logic.GC(w)
	if err != nil {
		errorPage(w, "ERROR:GC", err, 500)
		return
	}
}

func errorPage(w http.ResponseWriter, t string, e error, num int) {

	desc := fmt.Sprintf("%+v", e)

	log.Println(desc)

	dto := struct {
		Title   string
		Message string
		No      int
	}{t, desc, num}

	w.WriteHeader(num)

	viewManage(w, "error.tmpl", dto)
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
