package handler

import (
	"app/handler/manage"

	"net/http"

	"github.com/gorilla/mux"
)

func Register() {

	fs := http.FileServer(http.Dir("cmd/public"))
	http.Handle("/manage/js/", fs)
	http.Handle("/manage/css/", fs)

	mr := mux.NewRouter()
	h := manage.NewHandler(mr)

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

	//外部アクセス
	r := mux.NewRouter()
	pub := Public{r}

	r.HandleFunc("/page/{key}", pub.pageHandler).Methods("GET")
	r.HandleFunc("/file/{key}", pub.fileHandler).Methods("GET")
	r.HandleFunc("/file/{date}/{key}", pub.fileDateCacheHandler).Methods("GET")
	r.HandleFunc("/file_cache/{key}", pub.fileCacheHandler).Methods("GET")

	r.HandleFunc("/login", pub.loginHandler).Methods("GET")
	r.HandleFunc("/session", pub.sessionHandler).Methods("POST")
	r.HandleFunc("/sitemap/", pub.sitemap).Methods("GET")
	r.HandleFunc("/", pub.topHandler).Methods("GET")

	http.Handle("/", pub)
}
