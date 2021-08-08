package manage

import (
	"app/datastore"

	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func viewVariableHandler(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	cursor := q.Get("cursor")

	data, next, err := datastore.SelectVariables(r.Context(), cursor)
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
	id := ""
	//POST
	if POST(r) {

		r.ParseForm()
		id = r.FormValue("keyValue")
		check := r.FormValue("check")
		val := r.FormValue("variableData")
		ver := r.FormValue("version")

		if check == "true" {
			vari, err := datastore.SelectVariable(ctx, id)
			if err == nil && vari != nil {
				//存在する場合
				errorPage(w, "Error Exists Variable", err, 500)
				return
			}
			ver = "1"
		}

		//更新
		err := datastore.PutVariable(ctx, id, val, ver)
		if err != nil {
			errorPage(w, "Error Put Variable", err, 500)
			return
		}

		//TODO ここでリダイレクトかな、、、
	} else {
		vars := mux.Vars(r)
		id = vars["key"]
	}

	vari, err := datastore.SelectVariable(ctx, id)
	if err != nil {
		errorPage(w, "Error SelectVariable", err, 500)
		return
	}
	if vari == nil {
		errorPage(w, "NotFound Variable", fmt.Errorf(id), 404)
		return
	}

	variData, err := datastore.SelectVariableData(ctx, id)
	if err != nil {
		errorPage(w, "Not Found Variable Data", err, 500)
		return
	}
	if variData == nil {
		errorPage(w, "NotFound Variable Data", fmt.Errorf(id), 404)
		return
	}

	dto := struct {
		Variable     *datastore.Variable
		VariableData *datastore.VariableData
	}{vari, variData}
	viewManage(w, "variable/edit.tmpl", dto)
}

func deleteVariableHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	ctx := r.Context()

	err := datastore.RemoveVariable(ctx, id)
	if err != nil {
		errorPage(w, "Remove Template Error", err, 500)
		return
	}
	//リダイレクト
	http.Redirect(w, r, "/manage/variable/", 302)
}
