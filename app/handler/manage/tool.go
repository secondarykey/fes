package manage

import (
	"app/datastore"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func childrenPageHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	children, _, err := dao.SelectChildPages(ctx, id, datastore.NoLimitCursor, 0, true)
	if err != nil {
		errorPage(w, "Error Select Children page", err, 500)
		return
	}

	dto := struct {
		Parent string
		Pages  []datastore.Page
	}{id, children}
	viewManage(w, "page/children.tmpl", dto)
}

func changeSequencePageHandler(w http.ResponseWriter, r *http.Request) {

	id := r.FormValue("id")
	idCsv := r.FormValue("ids")
	enablesCsv := r.FormValue("enables")
	versionsCsv := r.FormValue("versions")

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	err := dao.PutPageSequence(ctx, idCsv, enablesCsv, versionsCsv)
	if err != nil {
		errorPage(w, "Error Page sequence update", err, 500)
		return
	}

	http.Redirect(w, r, "/manage/page/children/"+id, 302)
}

func referencePageTemplateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]
	referenceTemplateView(w, r, id, 2)
}

func referenceSiteTemplateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]
	referenceTemplateView(w, r, id, 1)
}

func referenceTemplateView(w http.ResponseWriter, r *http.Request, id string, typ int) {

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	pages, err := dao.SelectReferencePages(ctx, id, typ)
	if err != nil {
		errorPage(w, "Reference template pages Error", err, 500)
		return
	}
	if err != nil {
		errorPage(w, "Reference template pages Error", err, 500)
		return
	}
	if pages == nil || len(pages) <= 0 {
		errorPage(w, "Reference template pages NotFound", fmt.Errorf(id), 404)
		return
	}

	redirect := "/manage/template/edit/" + id

	dto := struct {
		ID       string
		Pages    []datastore.Page
		Redirect string
	}{id, pages, redirect}

	viewManage(w, "page/htmls.tmpl", dto)
}

type Tree struct {
	Page     *datastore.Page
	Children []*datastore.Tree
}

func treePageHandler(w http.ResponseWriter, r *http.Request) {

	dao := datastore.NewDao()
	defer dao.Close()

	tree, err := dao.CreatePagesTree(r.Context())
	if err != nil {
		errorPage(w, "Error Page Tree", err, 500)
		return
	}

	dto := struct {
		Tree *datastore.Tree
	}{tree}

	viewManage(w, "page/tree.tmpl", dto)
}
