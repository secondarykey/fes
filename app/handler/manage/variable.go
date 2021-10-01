package manage

import (
	"app/datastore"
	"app/handler/manage/form"

	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func viewVariableHandler(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	cursor := q.Get("cursor")

	dao := datastore.NewDao()
	defer dao.Close()

	data, next, err := dao.SelectVariables(r.Context(), cursor)
	if err != nil {
		errorPage(w, "Error Select Variables", err, 500)
		return
	}

	if data == nil {
		data = make([]datastore.Variable, 0)
	}

	dto := struct {
		Variables []datastore.Variable
		Now       string
		Next      string
	}{data, cursor, next}

	viewManage(w, "variable/view.tmpl", dto)
}

func addVariableHandler(w http.ResponseWriter, r *http.Request) {

	vari := &datastore.Variable{}
	variData := &datastore.VariableData{}

	//新規作成用のテンプレート
	dto := struct {
		Variable     *datastore.Variable
		VariableData *datastore.VariableData
	}{vari, variData}

	viewManage(w, "variable/edit.tmpl", dto)
}

func editVariableHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("X-XSS-Protection", "1")

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	vars := mux.Vars(r)
	id := vars["key"]

	var vari *datastore.Variable
	var variData *datastore.VariableData
	var err error

	r.ParseForm()
	vs := new(datastore.VariableSet)
	vs.ID = id
	if id == "" {
		vs.ID = r.FormValue("keyValue")
	}

	//指定がある場合
	vari, err = dao.SelectVariable(ctx, vs.ID)
	if err != nil {
		errorPage(w, "Error SelectVariable", err, 500)
		return
	}

	variData, err = dao.SelectVariableData(ctx, vs.ID)
	if err != nil {
		errorPage(w, "Not Found Variable Data", err, 500)
		return
	}

	check := r.FormValue("check")
	if check == "true" {
		if vari != nil {
			errorPage(w, "Exists Variable", fmt.Errorf("exists variable key=%s", vs.ID), 500)
			return
		}
	}

	if vari == nil {
		vari = new(datastore.Variable)
	}
	if variData == nil {
		variData = new(datastore.VariableData)
	}

	vs.Variable = vari
	vs.VariableData = variData

	//POST
	if POST(r) {
		err = form.SetVariable(r, vs)
		if err != nil {
			errorPage(w, "Error Put Variable", err, 500)
			return
		}
		//更新
		err := dao.PutVariable(ctx, vs)
		if err != nil {
			errorPage(w, "Error Put Variable", err, 500)
			return
		}
	}

	dto := struct {
		Variable     *datastore.Variable
		VariableData *datastore.VariableData
	}{vari, variData}

	viewManage(w, "variable/edit.tmpl", dto)
}

func uploadVariableHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	upload, header, err := r.FormFile("targetFile")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer upload.Close()

	ct := header.Header["Content-Type"]

	b, err := io.ReadAll(upload)
	if err != nil {
		fmt.Println(err)
		return
	}

	html := base64.StdEncoding.EncodeToString(b)

	if len(ct) > 0 {
		html = `<img src="data:` + ct[0] + ";base64," + html + `">`
	}

	w.Write([]byte(html))
}

func deleteVariableHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	err := dao.RemoveVariable(ctx, id)
	if err != nil {
		errorPage(w, "Remove Template Error", err, 500)
		return
	}
	//リダイレクト
	http.Redirect(w, r, "/manage/variable/", 302)
}
