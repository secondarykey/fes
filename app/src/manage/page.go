package manage

import (
	"datastore"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
)

func (h Handler) ViewPage(w http.ResponseWriter, r *http.Request) {
	h.view(w, r, "", "")
}

func (h Handler) AddPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parent := vars["key"]
	h.view(w, r, "", parent)
}

func (h Handler) EditPage(w http.ResponseWriter, r *http.Request) {

	if POST(r) {
		err := datastore.PutPage(r)
		if err != nil {
			h.errorPage(w, "Error Put Page",err.Error() ,500)
			return
		}
	}

	vars := mux.Vars(r)
	id := vars["key"]
	h.view(w, r, id, "")
}

func (h Handler) view(w http.ResponseWriter, r *http.Request, id string, parent string) {

	var err error
	var page *datastore.Page
	var pageData *datastore.PageData

	publish := false

	templates, err := datastore.SelectTemplates(r,-1)
	if err != nil {
		h.errorPage(w, "Error Select Template",err.Error(), 500)
		return
	}

	if id == "" && parent == "" {
		//親が空で検索
		page, err = datastore.SelectRootPage(r)
	} else if id != "" {
		page, err = datastore.SelectPage(r, id)
	}

	if err != nil {
		h.errorPage(w, "Error Select Page",err.Error() ,500)
		return
	}

	var children []datastore.Page
	siteTemplateName := "Select Site Template..."
	pageTemplateName := "Select Page Template..."

	if page == nil {

		//新規ページなので
		page = &datastore.Page{
			Parent: parent,
		}
		page.Deleted = true
		pageData = &datastore.PageData{}

		uid, err := uuid.NewV4()
		if err != nil {
			h.errorPage(w, "Generate uuid ",err.Error() ,500)
			return
		}

		id = uid.String()
		page.SetKey(datastore.CreatePageKey(r, id))
		pageData.SetKey(datastore.CreatePageDataKey(r, id))

		children = make([]datastore.Page, 0)

	} else {
		pageData, err = datastore.SelectPageData(r, page.Key.StringID())
		if err != nil {
			h.errorPage(w, "Error Select PageData", err.Error(),500)
			return
		}

		//全件でOK
		children, err = datastore.SelectChildPages(r, page.Key.StringID(),0,0,true)
		if err != nil {
			h.errorPage(w, "Error Select Children page", err.Error(),500)
			return
		}

		if !page.Deleted {
			if page.UpdatedAt.Unix() > page.Publish.Unix() + 5 {
				publish = true
			}
		}

		for _, elm := range templates {
			if elm.Key.StringID() == page.SiteTemplate {
				siteTemplateName = elm.Name
			}
			if elm.Key.StringID() == page.PageTemplate {
				pageTemplateName = elm.Name
			}
		}
	}

	wk := make([]datastore.Page, 0)
	wk = append(wk, *page)
	parent = page.Parent

	for {
		if parent == "" {
			break
		}
		parentPage, err := datastore.SelectPage(r, parent)
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
	}{page, pageData, children, breadcrumbs, templates, publish,siteTemplateName, pageTemplateName}
	h.parse(w, TEMPLATE_DIR+"page/edit.tmpl", dto)
}



func (h Handler) DeletePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	err := datastore.RemovePage(r, id)
	if err != nil {
		h.errorPage(w, "Error Delete Page", err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/manage/page/", 302)
}

func (h Handler) PublicPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]
	err := datastore.PutHTML(r,id)
	if err != nil {
		h.errorPage(w, "Error Private HTML", err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/manage/page/" + id, 302)
}

func (h Handler) PrivatePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]
	err := datastore.RemoveHTML(r,id)
	if err != nil {
		h.errorPage(w, "Error Private HTML", err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/manage/page/" + id, 302)
}

func (h Handler) ToolPage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]

	children, err := datastore.SelectChildPages(r, id,0,0,true)
	if err != nil {
		h.errorPage(w, "Error Select Children page", err.Error(),500)
		return
	}

	dto := struct {
		Parent string
		Pages []datastore.Page
	} {id,children}
	h.parse(w, TEMPLATE_DIR+"page/view.tmpl", dto)
}

func (h Handler) SequencePage(w http.ResponseWriter, r *http.Request) {

	id := r.FormValue("id")
	idCsv := r.FormValue("ids")
	enablesCsv := r.FormValue("enables")

	err := datastore.PutPageSequence(r,idCsv,enablesCsv)
	if err != nil {
		h.errorPage(w, "Error Page sequence update", err.Error(),500)
		return
	}

	http.Redirect(w, r, "/manage/page/tool/" + id, 302)
}

type Tree struct {
	Page     *datastore.Page
	Children []*datastore.Tree
}

func (h Handler) TreePage(w http.ResponseWriter, r *http.Request) {

	tree,err := datastore.PageTree(r)
	if err != nil {
		h.errorPage(w, "Error Page Tree", err.Error(), 500)
		return
	}

	dto := struct {
		Tree *datastore.Tree
	} {tree}

	h.parse(w, TEMPLATE_DIR+"page/tree.tmpl", dto)
}

