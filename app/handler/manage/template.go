package manage

import (
	"app/datastore"

	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (h Handler) ViewTemplate(w http.ResponseWriter, r *http.Request) {

	p := 1
	q := r.URL.Query()
	pageBuf := q.Get("page")
	if pageBuf != "" {
		page, err := strconv.Atoi(pageBuf)
		if err == nil {
			p = page
		}
	}

	data, err := datastore.SelectTemplates(r, p)
	if err != nil {
		h.errorPage(w, "Error Select Template", err.Error(), 500)
		return
	}

	if data == nil {
		data = make([]datastore.Template, 0)
	}

	dto := struct {
		Templates []datastore.Template
		Page      int
		Prev      int
		Next      int
	}{data, p, p - 1, p + 1}

	h.parse(w, TEMPLATE_DIR+"template/view.tmpl", dto)
}

func (h Handler) AddTemplate(w http.ResponseWriter, r *http.Request) {
	tmp := &datastore.Template{}
	tmpData := &datastore.TemplateData{}
	tmp.LoadKey(datastore.CreateTemplateKey())
	//新規作成用のテンプレート
	dto := struct {
		Template     *datastore.Template
		TemplateData *datastore.TemplateData
	}{tmp, tmpData}
	h.parse(w, TEMPLATE_DIR+"template/edit.tmpl", dto)
}

func (h Handler) EditTemplate(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("X-XSS-Protection", "1")
	//POST
	if POST(r) {
		//更新
		err := datastore.PutTemplate(r)
		if err != nil {
			h.errorPage(w, "Error Put Template", err.Error(), 500)
			return
		}
	}
	vars := mux.Vars(r)
	id := vars["key"]
	tmp, err := datastore.SelectTemplate(r, id)
	if err != nil {
		h.errorPage(w, "Error SelectTemplate", err.Error(), 500)
		return
	}
	if tmp == nil {
		h.errorPage(w, "NotFound Template", id, 404)
		return
	}

	tmpData, err := datastore.SelectTemplateData(r, id)
	if err != nil {
		h.errorPage(w, err.Error(), id, 500)
		return
	}
	if tmpData == nil {
		h.errorPage(w, "NotFound TemplateData", id, 404)
		return
	}

	dto := struct {
		Template     *datastore.Template
		TemplateData *datastore.TemplateData
	}{tmp, tmpData}
	h.parse(w, TEMPLATE_DIR+"template/edit.tmpl", dto)
}

func (h Handler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	if datastore.UsingTemplate(r, id) {
		h.errorPage(w, "Using Template", id, 500)
		return
	}

	err := datastore.RemoveTemplate(r, id)
	if err != nil {
		h.errorPage(w, err.Error(), id, 500)
		return
	}
	//リダイレクト
	http.Redirect(w, r, "/manage/template/", 302)
}

func (h Handler) ReferenceTemplate(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]

	t := r.FormValue("type")
	//参照しているページを取得
	pages, err := datastore.SelectReferencePages(r, id, t)

	if err != nil {
		h.errorPage(w, "Reference template pages Error", err.Error(), 500)
		return
	}
	if pages == nil || len(pages) <= 0 {
		h.errorPage(w, "Reference template pages NotFound", id, 404)
		return
	}

	//ページからHTMLを更新
	err = datastore.PutHTMLs(r, pages)
	if err != nil {
		h.errorPage(w, "Put HTML data Error", err.Error(), 500)
		return
	}

	http.Redirect(w, r, "/manage/template/edit/"+id, 302)
}
