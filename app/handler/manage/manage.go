package manage

import (
	"app/datastore"
	. "app/handler/internal"
	"app/logic"

	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

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

	// V1 ルートは /manage/v1/ 配下
	v1 := s.PathPrefix("/v1").Subrouter()

	v1.HandleFunc("/favico.ico", viewRootPageHandler).Methods("GET")

	//Page
	v1.HandleFunc("/page/", viewRootPageHandler).Methods("GET")
	v1.HandleFunc("/page/{key}", viewPageHandler)
	v1.HandleFunc("/page/add/{key}", addPageHandler).Methods("GET")
	v1.HandleFunc("/page/delete/{key}", deletePageHandler).Methods("GET")

	//Tool
	v1.HandleFunc("/page/children/{key}", childrenPageHandler).Methods("GET")
	v1.HandleFunc("/page/update/sequence", changeSequencePageHandler).Methods("POST")
	v1.HandleFunc("/page/update/sort", changeSortPageHandler).Methods("POST")
	v1.HandleFunc("/page/update/move", movePageHandler).Methods("POST")
	v1.HandleFunc("/page/template/page/{key}", referencePageTemplateHandler).Methods("GET")
	v1.HandleFunc("/page/template/site/{key}", referenceSiteTemplateHandler).Methods("GET")
	v1.HandleFunc("/page/tree/", treePageHandler).Methods("GET")

	//HTML
	v1.HandleFunc("/html/publish/{key}", changePublishPageHandler).Methods("GET")
	v1.HandleFunc("/html/private/{key}", changePrivatePageHandler).Methods("GET")
	v1.HandleFunc("/html/update", updateHTMLHandler).Methods("POST")

	//ページ表示（V1互換）
	v1.HandleFunc("/page/view/{key}", privatePageHandler).Methods("GET")
	v1.HandleFunc("/page/view/", privateHandler).Methods("GET")

	//Draft
	v1.HandleFunc("/draft/", viewDraftHandler).Methods("GET")
	v1.HandleFunc("/draft/add", addDraftHandler).Methods("GET")
	v1.HandleFunc("/draft/publish/{key}", publishDraftHandler)
	v1.HandleFunc("/draft/current/{key}", currentDraftHandler)
	v1.HandleFunc("/draft/edit/{key}", editDraftHandler)
	v1.HandleFunc("/draft/delete/{key}", deleteDraftHandler)
	v1.HandleFunc("/draft/page/add/{key}", addDraftPageHandler)
	v1.HandleFunc("/draft/page/delete/{key}", deleteDraftPageHandler)

	//File
	v1.HandleFunc("/file/", viewFileHandler).Methods("GET")
	v1.HandleFunc("/file/type/{type}", viewFileHandler).Methods("GET")
	v1.HandleFunc("/file/add", addFileHandler).Methods("POST")
	v1.HandleFunc("/file/favicon", faviconUploadHandler).Methods("POST")
	v1.HandleFunc("/file/delete", deleteFileHandler).Methods("POST")
	v1.HandleFunc("/file/resize/{key}", resizeFileHandler).Methods("GET")
	v1.HandleFunc("/file/resize/commit", resizeCommitFileHandler).Methods("POST")
	v1.HandleFunc("/file/resize/view/{key}", resizeFileViewHandler).Methods("GET")
	v1.HandleFunc("/file/view/{key}", fileViewHandler).Methods("GET")

	//Template
	v1.HandleFunc("/template/", viewTemplateHandler).Methods("GET")
	v1.HandleFunc("/template/type/{type}", viewTemplateHandler).Methods("GET")
	v1.HandleFunc("/template/add", addTemplateHandler).Methods("GET")
	v1.HandleFunc("/template/edit/{key}", editTemplateHandler)
	v1.HandleFunc("/template/delete/{key}", deleteTemplateHandler)
	v1.HandleFunc("/template/reference/{type}/{key}", referenceTemplateHandler)

	v1.HandleFunc("/variable/", viewVariableHandler).Methods("GET")
	v1.HandleFunc("/variable/add", addVariableHandler).Methods("GET")
	v1.HandleFunc("/variable/edit", editVariableHandler)
	v1.HandleFunc("/variable/upload", uploadVariableHandler).Methods("POST")
	v1.HandleFunc("/variable/edit/{key}", editVariableHandler).Methods("GET")
	v1.HandleFunc("/variable/delete/{key}", deleteVariableHandler)

	//table
	v1.HandleFunc("/table/view", viewTableHandler)

	v1.HandleFunc("/datastore/backup", backupHandler).Methods("POST")
	v1.HandleFunc("/datastore/restore", restoreHandler).Methods("POST")

	//Site
	v1.HandleFunc("/site/", viewSiteHandler).Methods("GET")
	v1.HandleFunc("/site/edit", editSiteHandler).Methods("POST")
	v1.HandleFunc("/site/map", downloadSitemapHandler).Methods("GET")
	v1.HandleFunc("/site/clean", cleanSiteHandler).Methods("GET", "POST")

	v1.HandleFunc("/system/gc", gc).Methods("GET")

	v1.HandleFunc("/", indexHandler).Methods("GET")

	// ページプレビュー（/manage/page/view/）— SPA catch-all より前に登録
	s.HandleFunc("/page/view/{key}", privatePageHandler).Methods("GET")
	s.HandleFunc("/page/view/", privateHandler).Methods("GET")

	// SPA + API (/manage/api/ + /manage/ キャッチオール)
	if err := registerAPI(s); err != nil {
		return xerrors.Errorf("registerAPI() error: %w", err)
	}

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
		if isAPIPath(r.URL.Path) {
			apiUnauthorized(w)
			return
		}
		http.Redirect(w, r, "/login?redirect="+r.URL.RequestURI(), 302)
		return
	}

	if u == nil {
		log.Println("ユーザがいない")
		if isAPIPath(r.URL.Path) {
			apiUnauthorized(w)
			return
		}
		http.Redirect(w, r, "/login?redirect="+r.URL.RequestURI(), 302)
		return
	}

	h.r.ServeHTTP(w, r)
}

func isAPIPath(path string) bool {
	return strings.HasPrefix(path, "/manage/api/")
}

func apiUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
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
