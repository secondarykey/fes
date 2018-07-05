package src

import (
	"github.com/gorilla/mux"
	"manage"
	"net/http"
)

func init() {

	r := mux.NewRouter()

	h := manage.Handler{}

	r.HandleFunc("/manage/", h.View).Methods("GET")
	//Page
	r.HandleFunc("/manage/page/", h.ViewPage).Methods("GET")
	r.HandleFunc("/manage/page/{key}", h.EditPage)
	r.HandleFunc("/manage/page/add/{key}", h.AddPage).Methods("GET")
	r.HandleFunc("/manage/page/delete/{key}", h.DeletePage).Methods("GET")
	r.HandleFunc("/manage/page/public/{key}", h.PublicPage).Methods("GET")
	r.HandleFunc("/manage/page/private/{key}", h.PrivatePage).Methods("GET")
	r.HandleFunc("/manage/page/view/{key}", h.PageHandler).Methods("GET")
	r.HandleFunc("/manage/page/view/", h.TopHandler).Methods("GET")

	//File
	r.HandleFunc("/manage/file/", h.ViewFile).Methods("GET")
	r.HandleFunc("/manage/file/add", h.AddFile).Methods("POST")
	r.HandleFunc("/manage/file/delete", h.DeleteFile).Methods("POST")
	r.HandleFunc("/manage/file/resize/{key}", h.ResizeFile).Methods("GET")
	r.HandleFunc("/manage/file/resize/view/{key}", h.ResizeFileView).Methods("GET")

	//Template
	r.HandleFunc("/manage/template/", h.ViewTemplate).Methods("GET")
	r.HandleFunc("/manage/template/add", h.AddTemplate).Methods("GET")
	r.HandleFunc("/manage/template/edit/{key}", h.EditTemplate)
	r.HandleFunc("/manage/template/delete/{key}", h.DeleteTemplate)

	//Site
	r.HandleFunc("/manage/site/", h.ViewSetting).Methods("GET")
	r.HandleFunc("/manage/site/edit", h.EditSetting).Methods("POST")
	r.HandleFunc("/manage/site/map", h.DownloadSitemap).Methods("GET")

	//外部アクセス
	pub := Public{}
	r.HandleFunc("/page/{key}", pub.pageHandler).Methods("GET")
	r.HandleFunc("/file/{key}", pub.fileHandler).Methods("GET")
	r.HandleFunc("/", pub.topHandler).Methods("GET")

	http.Handle("/", r)
}
