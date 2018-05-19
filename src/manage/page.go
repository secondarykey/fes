package manage

import (
	"datastore"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
)

func (h Handler) ViewPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]
	h.view(w, r, id, "")
}

func (h Handler) AddPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parent := vars["key"]
	h.view(w, r, "", parent)
}

func (h Handler) view(w http.ResponseWriter, r *http.Request, id string, parent string) {

	var err error
	var page *datastore.Page
	var pageData *datastore.PageData

	templates, err := datastore.SelectTemplates(r,-1)
	if err != nil {
		h.errorPage(w, err.Error(), "Select template", 500)
		return
	}

	if id == "" && parent == "" {
		//親が空で検索
		page, err = datastore.SelectRootPage(r)
	} else if id != "" {
		page, err = datastore.SelectPage(r, id)
	}

	if err != nil {
		h.errorPage(w, err.Error(), "Select Page", 500)
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
		pageData = &datastore.PageData{}

		uid, err := uuid.NewV4()
		if err != nil {
		}

		id = uid.String()
		page.SetKey(datastore.CreatePageKey(r, id))
		pageData.SetKey(datastore.CreatePageDataKey(r, id))

		children = make([]datastore.Page, 0)

	} else {
		pageData, err = datastore.SelectPageData(r, page.Key.StringID())
		if err != nil {
			h.errorPage(w, err.Error(), "Select PageData", 500)
			return
		}

		children, err = datastore.SelectChildPages(r, page.Key.StringID(),0)
		if err != nil {
			h.errorPage(w, err.Error(), "Children page", 500)
			return
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
		SiteTemplateName string
		PageTemplateName string
	}{page, pageData, children, breadcrumbs, templates, siteTemplateName, pageTemplateName}
	h.parse(w, TEMPLATE_DIR+"page/edit.tmpl", dto)
}

func (h Handler) EditPage(w http.ResponseWriter, r *http.Request) {

	if POST(r) {
		err := datastore.PutPage(r)
		if err != nil {
			h.errorPage(w, err.Error(), "Put Page", 500)
			return
		}
	}

	vars := mux.Vars(r)
	id := vars["key"]
	h.view(w, r, id, "")
}

func (h Handler) DeletePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	err := datastore.RemovePage(r, id)
	if err != nil {
		h.errorPage(w, "Delete Page", err.Error(), 500)
	}
	http.Redirect(w, r, "/manage/page/", 302)
}
