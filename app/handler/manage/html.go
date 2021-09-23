package manage

import (
	"app/datastore"
	"app/logic"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func updateHTMLHandler(w http.ResponseWriter, r *http.Request) {

	redirect := r.FormValue("redirect")
	//IDsをそのままPublish
	idcsv := r.FormValue("ids")
	ids := strings.Split(idcsv, ",")

	ctx := r.Context()

	err := logic.PutHTMLs(ctx, ids...)
	if err != nil {
		errorPage(w, "Error SelectPages() error", err, 500)
		return
	}

	http.Redirect(w, r, redirect, 302)
}

func changePublishPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	ctx := r.Context()

	err := logic.PutHTMLs(ctx, id)
	if err != nil {
		errorPage(w, "Error Publish HTML", err, 500)
		return
	}
	http.Redirect(w, r, "/manage/page/"+id, 302)
}

func changePrivatePageHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]

	ctx := r.Context()

	dao := datastore.NewDao()
	defer dao.Close()

	err := dao.RemoveHTML(ctx, id)
	if err != nil {
		errorPage(w, "Error Private HTML", err, 500)
		return
	}
	http.Redirect(w, r, "/manage/page/"+id, 302)
}
