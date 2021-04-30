package manage

import (
	"app/datastore"
	"app/logic"

	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
)

func viewTemplateHandler(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	cursor := q.Get("cursor")

	data, next, err := datastore.SelectTemplates(r, cursor)
	if err != nil {
		errorPage(w, "Error Select Template", err, 500)
		return
	}

	if data == nil {
		data = make([]datastore.Template, 0)
	}

	dto := struct {
		Templates []datastore.Template
		Now       string
		Next      string
	}{data, cursor, next}

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

	ctx := r.Context()

	if ok, err := datastore.UsingTemplate(ctx, id); err != nil {
		errorPage(w, "Using Template", xerrors.Errorf("datastore.UsingTemplate() error : %w", err), 500)
		return
	} else if !ok {
		errorPage(w, "Using Template", fmt.Errorf("Using template[%s]", id), 500)
		return
	}

	err := datastore.RemoveTemplate(ctx, id)
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

	fmt.Println(len(pages))

	//ページからHTMLを更新
	err = logic.PutHTMLs(r, pages)
	if err != nil {
		errorPage(w, "Put HTML data Error", err, 500)
		return
	}

	http.Redirect(w, r, "/manage/template/edit/"+id, 302)
}
