package manage

import (
	"net/http"
	"datastore"
	"api"

	"strconv"
)

//URL = /manage/file/
func (h Handler) ViewFile(w http.ResponseWriter, r *http.Request) {

	p:= 1
	q := r.URL.Query()
	pageBuf := q.Get("page")
	if pageBuf != "" {
		page,err := strconv.Atoi(pageBuf)
		if err == nil {
			p = page
		}
	}

	files,err := datastore.SelectFiles(r,p)
	if err != nil {
		h.errorPage(w,err.Error(),"Select File",500)
		return
	}

	dto := struct {
		Files []datastore.File
		Page int
		Prev int
		Next int
	} {files,p,p-1,p+1}
	h.parse(w, TEMPLATE_DIR + "file/view.tmpl", dto)
}

//URL = /manage/file/add
func (h Handler) AddFile(w http.ResponseWriter, r *http.Request) {

	err := datastore.SaveFile(r,"",api.DATA_FILE)
	if err != nil {
		h.errorPage(w,err.Error(),"Add File Error",500)
		return
	}
	//リダイレクト
	http.Redirect(w, r, "/manage/file/", 302)
}

//URL = /manage/file/delete
func (h Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	//リダイレクト
	id := r.FormValue("fileName")
	err := datastore.RemoveFile(r,id)
	if err != nil {
		h.errorPage(w,err.Error(),id,500)
	}
	http.Redirect(w, r, "/manage/file/", 302)
}
