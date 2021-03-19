package manage

import (
	"app/datastore"
	"app/logic"

	"net/http"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
)

func viewRootPageHandler(w http.ResponseWriter, r *http.Request) {

	page, err := datastore.SelectRootPage(r)
	if err != nil {
		if err == datastore.SiteNotFoundError {
			http.Redirect(w, r, "/manage/site/", 302)
		} else {
			errorPage(w, "Select Root Page error", err, 500)
		}
		return
	}
	view(w, r, page)
}

func addPageHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	parent := vars["key"]

	//新規ページなので
	page := &datastore.Page{
		Parent: parent,
	}
	page.Deleted = true

	uid := uuid.NewV4()

	id := uid.String()
	page.LoadKey(datastore.CreatePageKey(id))

	view(w, r, page)
}

func viewPageHandler(w http.ResponseWriter, r *http.Request) {

	if POST(r) {
		err := datastore.PutPage(r)
		if err != nil {
			errorPage(w, "Error Put Page", err, 500)
			return
		}
	}

	vars := mux.Vars(r)
	id := vars["key"]
	//ページ検索
	page, err := datastore.SelectPage(r, id, -1)
	if err != nil {
		errorPage(w, "Error Select Page", err, 500)
		return
	}

	view(w, r, page)
}

func view(w http.ResponseWriter, r *http.Request, page *datastore.Page) {

	var err error
	var pageData *datastore.PageData

	publish := false

	//全件検索
	templates, _, err := datastore.SelectTemplates(r, datastore.NoLimitCursor)
	if err != nil {
		errorPage(w, "Error Select Template", err, 500)
		return
	}

	var children []datastore.Page
	siteTemplateName := "Select Site Template..."
	pageTemplateName := "Select Page Template..."

	id := page.Key.Name

	pageData, err = datastore.SelectPageData(r, id)
	if err != nil {
		errorPage(w, "Error Select PageData", err, 500)
		return
	}

	if pageData == nil {
		pageData = &datastore.PageData{}
		pageData.LoadKey(datastore.CreatePageDataKey(id))
	}

	//全件でOK
	children, _, err = datastore.SelectChildPages(r, id, datastore.NoLimitCursor, 0, true)
	if err != nil {
		errorPage(w, "Error Select Children page", err, 500)
		return
	}

	if children == nil {
		children = make([]datastore.Page, 0)
	}

	if !page.Deleted {
		if page.UpdatedAt.Unix() > page.Publish.Unix()+5 {
			publish = true
		}
	}

	for _, elm := range templates {
		if elm.Key.Name == page.SiteTemplate {
			siteTemplateName = elm.Name
		}
		if elm.Key.Name == page.PageTemplate {
			pageTemplateName = elm.Name
		}
	}

	wk := make([]datastore.Page, 0)
	wk = append(wk, *page)

	parent := page.Parent

	for {
		if parent == "" {
			break
		}
		parentPage, err := datastore.SelectPage(r, parent, -1)
		if err != nil {
			break
		}
		wk = append(wk, *parentPage)
		parent = parentPage.Parent
	}

	breadcrumbs := make([]datastore.Page, len(wk))
	for idx, _ := range wk {
		breadcrumbs[idx] = wk[len(wk)-1-idx]
	}

	dto := struct {
		Page             *datastore.Page
		PageData         *datastore.PageData
		Children         []datastore.Page
		Breadcrumbs      []datastore.Page
		Templates        []datastore.Template
		Publish          bool
		SiteTemplateName string
		PageTemplateName string
	}{page, pageData, children, breadcrumbs, templates, publish, siteTemplateName, pageTemplateName}

	viewManage(w, "page/edit.tmpl", dto)
}

func deletePageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	err := datastore.RemovePage(r, id)
	if err != nil {
		errorPage(w, "Error Delete Page", err, 500)
		return
	}
	http.Redirect(w, r, "/manage/page/", 302)
}

func changePublicPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	err := logic.PutHTML(r, id)
	if err != nil {
		errorPage(w, "Error Publish HTML", err, 500)
		return
	}
	http.Redirect(w, r, "/manage/page/"+id, 302)
}

func changePrivatePageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]
	err := datastore.RemoveHTML(r, id)
	if err != nil {
		errorPage(w, "Error Private HTML", err, 500)
		return
	}
	http.Redirect(w, r, "/manage/page/"+id, 302)
}

func toolPageHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]

	children, _, err := datastore.SelectChildPages(r, id, datastore.NoLimitCursor, 0, true)
	if err != nil {
		errorPage(w, "Error Select Children page", err, 500)
		return
	}

	dto := struct {
		Parent string
		Pages  []datastore.Page
	}{id, children}
	viewManage(w, "page/tool.tmpl", dto)
}

func changeSequencePageHandler(w http.ResponseWriter, r *http.Request) {

	id := r.FormValue("id")
	idCsv := r.FormValue("ids")
	enablesCsv := r.FormValue("enables")
	versionsCsv := r.FormValue("versions")

	err := datastore.PutPageSequence(r, idCsv, enablesCsv, versionsCsv)
	if err != nil {
		errorPage(w, "Error Page sequence update", err, 500)
		return
	}

	http.Redirect(w, r, "/manage/page/tool/"+id, 302)
}

type Tree struct {
	Page     *datastore.Page
	Children []*datastore.Tree
}

func treePageHandler(w http.ResponseWriter, r *http.Request) {

	tree, err := datastore.PageTree(r.Context())
	if err != nil {
		errorPage(w, "Error Page Tree", err, 500)
		return
	}

	dto := struct {
		Tree *datastore.Tree
	}{tree}

	viewManage(w, "page/tree.tmpl", dto)
}
