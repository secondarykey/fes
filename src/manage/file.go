package manage

import (
	"net/http"
	"datastore"

)

func (h Handler) ViewFile(w http.ResponseWriter, r *http.Request) {

	files,err := datastore.SelectFiles(r)
	if err != nil {
		h.errorPage(w,err.Error(),"Select File",500)
		return
	}

	dto := struct {
		Files []datastore.File
	} {files}
	h.parse(w, TEMPLATE_DIR + "file/view.tmpl", dto)
}

func (h Handler) AddFile(w http.ResponseWriter, r *http.Request) {

	err := datastore.SaveFile(r,"")
	if err != nil {
		h.errorPage(w,err.Error(),"Add File Error",500)
		return
	}
	//リダイレクト
	http.Redirect(w, r, "/manage/file/", 302)
}

func (h Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	//リダイレクト
	id := r.FormValue("fileName")
	err := datastore.RemoveFile(r,id)
	if err != nil {
		h.errorPage(w,err.Error(),id,500)
	}
	http.Redirect(w, r, "/manage/file/", 302)
}
