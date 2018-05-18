package manage

import (
	"net/http"

	"datastore"

	"github.com/gorilla/mux"
)

func (h Handler) ViewTemplate(w http.ResponseWriter, r *http.Request) {

	data,err := datastore.SelectTemplates(r)
	if err != nil {
		h.errorPage(w,err.Error(),"Select Error",500)
		return
	}

	if data == nil {
		data = make([]datastore.Template,0)
	}

	dto := struct {
		Templates []datastore.Template
	} {data}

	h.parse(w, TEMPLATE_DIR + "template/view.tmpl", dto)
}

func (h Handler) AddTemplate(w http.ResponseWriter, r *http.Request) {

	tmp := &datastore.Template{}
	tmpData := &datastore.TemplateData{}

	tmp.SetKey(datastore.CreateTemplateKey(r))

	//新規作成用のテンプレート
	dto := struct {
		Template *datastore.Template
		TemplateData *datastore.TemplateData
	} {tmp,tmpData}
	h.parse(w, TEMPLATE_DIR + "template/edit.tmpl", dto)
}

func (h Handler) EditTemplate(w http.ResponseWriter, r *http.Request) {

	//POST
	if POST(r) {
		//更新
		err := datastore.PutTemplate(r)
		if err != nil {

		}
		//JSONで返す

	} else {
		vars := mux.Vars(r)
		id := vars["key"]
		tmp,err := datastore.SelectTemplate(r,id)
		if err != nil {
			h.errorPage(w,err.Error(),id,500)
			return
		}
		if tmp == nil {
			h.errorPage(w,"NotFound Template",id ,500)
			return
		}

		tmpData ,err := datastore.SelectTemplateData(r,id)
		if err != nil {
			h.errorPage(w,err.Error(),id,500)
			return
		}
		if tmpData == nil {
			h.errorPage(w,"NotFound TemplateData",id,500)
			return
		}

		dto := struct {
			Template *datastore.Template
			TemplateData *datastore.TemplateData
		} {tmp,tmpData}
		h.parse(w, TEMPLATE_DIR + "template/edit.tmpl", dto)
	}
}

func (h Handler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	if datastore.UsingTemplate(r,id) {
		h.errorPage(w,"Using Template",id,500)
		return
	}

	err := datastore.RemoveTemplate(r,id)
	if err != nil {
		h.errorPage(w,err.Error(),id,500)
		return
	}
	//リダイレクト
	http.Redirect(w, r, "/manage/template/", 302)
}
