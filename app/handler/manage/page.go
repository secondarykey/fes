package manage

import (
	"app/datastore"

	"net/http"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
)

func (h Handler) ViewRootPage(w http.ResponseWriter, r *http.Request) {

	page, err := datastore.SelectRootPage(r)
	if err != nil {
		if err == datastore.SiteNotFoundError {
			http.Redirect(w, r, "/manage/site/", 302)
		} else {
			h.errorPage(w, "Select Root Page error", err, 500)
		}
		return
	}
	h.view(w, r, page)
}

func (h Handler) AddPage(w http.ResponseWriter, r *http.Request) {

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

	h.view(w, r, page)
}

func (h Handler) ViewPage(w http.ResponseWriter, r *http.Request) {

	if POST(r) {
		err := datastore.PutPage(r)
		if err != nil {
			h.errorPage(w, "Error Put Page", err, 500)
			return
		}
	}

	vars := mux.Vars(r)
	id := vars["key"]
	//ページ検索
	page, err := datastore.SelectPage(r, id, -1)
	if err != nil {
		h.errorPage(w, "Error Select Page", err, 500)
		return
	}

	h.view(w, r, page)
}

func (h Handler) view(w http.ResponseWriter, r *http.Request, page *datastore.Page) {

	var err error
	var pageData *datastore.PageData

	publish := false

	//全件検索
	templates, err := datastore.SelectTemplates(r, -1)
	if err != nil {
		h.errorPage(w, "Error Select Template", err, 500)
		return
	}

	var children []datastore.Page
	siteTemplateName := "Select Site Template..."
	pageTemplateName := "Select Page Template..."

	id := page.Key.Name

	pageData, err = datastore.SelectPageData(r, id)
	if err != nil {
		h.errorPage(w, "Error Select PageData", err, 500)
		return
	}

	if pageData == nil {
		pageData = &datastore.PageData{}
		pageData.LoadKey(datastore.CreatePageDataKey(id))
	}

	//全件でOK
	children, err = datastore.SelectChildPages(r, id, 0, 0, true)
	if err != nil {
		h.errorPage(w, "Error Select Children page", err, 500)
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
	h.parse(w, TEMPLATE_DIR+"page/edit.tmpl", dto)
}

func (h Handler) DeletePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	err := datastore.RemovePage(r, id)
	if err != nil {
		h.errorPage(w, "Error Delete Page", err, 500)
		return
	}
	http.Redirect(w, r, "/manage/page/", 302)
}

func (h Handler) PublicPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	err := datastore.PutHTML(r, id)
	if err != nil {
		h.errorPage(w, "Error Publish HTML", err, 500)
		return
	}
	http.Redirect(w, r, "/manage/page/"+id, 302)
}

func (h Handler) PrivatePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]
	err := datastore.RemoveHTML(r, id)
	if err != nil {
		h.errorPage(w, "Error Private HTML", err, 500)
		return
	}
	http.Redirect(w, r, "/manage/page/"+id, 302)
}

func (h Handler) ToolPage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]

	children, err := datastore.SelectChildPages(r, id, 0, 0, true)
	if err != nil {
		h.errorPage(w, "Error Select Children page", err, 500)
		return
	}

	dto := struct {
		Parent string
		Pages  []datastore.Page
	}{id, children}
	h.parse(w, TEMPLATE_DIR+"page/tool.tmpl", dto)
}

func (h Handler) SequencePage(w http.ResponseWriter, r *http.Request) {

	id := r.FormValue("id")
	idCsv := r.FormValue("ids")
	enablesCsv := r.FormValue("enables")
	versionsCsv := r.FormValue("versions")

	err := datastore.PutPageSequence(r, idCsv, enablesCsv, versionsCsv)
	if err != nil {
		h.errorPage(w, "Error Page sequence update", err, 500)
		return
	}

	http.Redirect(w, r, "/manage/page/tool/"+id, 302)
}

type Tree struct {
	Page     *datastore.Page
	Children []*datastore.Tree
}

func (h Handler) TreePage(w http.ResponseWriter, r *http.Request) {

	tree, err := datastore.PageTree(r)
	if err != nil {
		h.errorPage(w, "Error Page Tree", err, 500)
		return
	}

	dto := struct {
		Tree *datastore.Tree
	}{tree}

	h.parse(w, TEMPLATE_DIR+"page/tree.tmpl", dto)
}
