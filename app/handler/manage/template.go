package manage

import (
	"app/datastore"
	"fmt"

	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func viewTemplateHandler(w http.ResponseWriter, r *http.Request) {

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
		errorPage(w, "Error Select Template", err, 500)
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

	viewManage(w, "template/view.tmpl", dto)
}

func addTemplateHandler(w http.ResponseWriter, r *http.Request) {
	tmp := &datastore.Template{}
	tmpData := &datastore.TemplateData{}
	tmp.LoadKey(datastore.CreateTemplateKey())
	//新規作成用のテンプレート
	dto := struct {
		Template     *datastore.Template
		TemplateData *datastore.TemplateData
	}{tmp, tmpData}

	viewManage(w, "template/edit.tmpl", dto)
}

func editTemplateHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("X-XSS-Protection", "1")
	//POST
	if POST(r) {
		//更新
		err := datastore.PutTemplate(r)
		if err != nil {
			errorPage(w, "Error Put Template", err, 500)
			return
		}
	}
	vars := mux.Vars(r)
	id := vars["key"]
	tmp, err := datastore.SelectTemplate(r, id)
	if err != nil {
		errorPage(w, "Error SelectTemplate", err, 500)
		return
	}
	if tmp == nil {
		errorPage(w, "NotFound Template", fmt.Errorf(id), 404)
		return
	}

	tmpData, err := datastore.SelectTemplateData(r, id)
	if err != nil {
		errorPage(w, "Not Found Template Data", err, 500)
		return
	}
	if tmpData == nil {
		errorPage(w, "NotFound Template Data", fmt.Errorf(id), 404)
		return
	}

	dto := struct {
		Template     *datastore.Template
		TemplateData *datastore.TemplateData
	}{tmp, tmpData}
	viewManage(w, "template/edit.tmpl", dto)
}

func deleteTemplateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	if datastore.UsingTemplate(r, id) {
		errorPage(w, "Using Template", fmt.Errorf(id), 500)
		return
	}

	err := datastore.RemoveTemplate(r, id)
	if err != nil {
		errorPage(w, "Remove Template Error", err, 500)
		return
	}
	//リダイレクト
	http.Redirect(w, r, "/manage/template/", 302)
}

func referenceTemplateHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]

	t := r.FormValue("type")
	//参照しているページを取得
	pages, err := datastore.SelectReferencePages(r, id, t)

	if err != nil {
		errorPage(w, "Reference template pages Error", err, 500)
		return
	}
	if pages == nil || len(pages) <= 0 {
		errorPage(w, "Reference template pages NotFound", fmt.Errorf(id), 404)
		return
	}

	//ページからHTMLを更新
	err = datastore.PutHTMLs(r, pages)
	if err != nil {
		errorPage(w, "Put HTML data Error", err, 500)
		return
	}

	http.Redirect(w, r, "/manage/template/edit/"+id, 302)
}
