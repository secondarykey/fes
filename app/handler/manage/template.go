package manage

import (
	"app/datastore"
	"app/handler/manage/form"

	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
)

func viewTemplateHandler(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	cursor := q.Get("cursor")

	vars := mux.Vars(r)
	t, flag := vars["type"]
	if !flag {
		t = "2"
	}

	dao := datastore.NewDao()
	defer dao.Close()

	data, next, err := dao.SelectTemplates(r.Context(), t, cursor)
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

	dao := datastore.NewDao()
	defer dao.Close()

	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["key"]

	//POST
	if POST(r) {

		tm, err := dao.SelectTemplate(ctx, id)
		if err != nil {
			errorPage(w, "Error SelectTemplate", err, 500)
			return
		}

		ts := new(datastore.TemplateSet)
		ts.Template = tm
		if tm == nil {
			ts.Template = new(datastore.Template)
		}
		ts.TemplateData = new(datastore.TemplateData)

		err = form.SetTemplate(r, ts)
		if err != nil {
			errorPage(w, "Error CreateFormTemplate()", err, 500)
			return
		}

		//更新
		err = dao.PutTemplate(ctx, ts)
		if err != nil {
			errorPage(w, "Error Put Template", err, 500)
			return
		}
	}

	tmp, err := dao.SelectTemplate(ctx, id)
	if err != nil {
		errorPage(w, "Error SelectTemplate", err, 500)
		return
	}
	if tmp == nil {
		errorPage(w, "NotFound Template", fmt.Errorf(id), 404)
		return
	}

	tmpData, err := dao.SelectTemplateData(ctx, id)
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
	dao := datastore.NewDao()
	defer dao.Close()

	if ok, err := dao.UsingTemplate(ctx, id); err != nil {
		errorPage(w, "Using Template", xerrors.Errorf("datastore.UsingTemplate() error : %w", err), 500)
		return
	} else if ok {
		errorPage(w, "Using Template", fmt.Errorf("Using template[%s]", id), 500)
		return
	}

	err := dao.RemoveTemplate(ctx, id)
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
	t := vars["type"]

	typ := "site"
	if t == "2" {
		typ = "page"
	}

	http.Redirect(w, r, "/manage/page/template/"+typ+"/"+id, 302)
}
