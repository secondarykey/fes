package manage

import (
	"app/datastore"
	. "app/handler/internal"
	"app/logic"

	"net/http"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
)

func viewRootPageHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	page, err := datastore.SelectRootPage(ctx)
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
	page := datastore.Page{}

	page.Parent = parent
	page.Deleted = true

	uid := uuid.NewV4()
	id := uid.String()

	page.LoadKey(datastore.CreatePageKey(id))

	view(w, r, &page)
}

func viewPageHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	if POST(r) {

		vars := mux.Vars(r)
		id := vars["key"]

		var ps datastore.PageSet

		p, pd, err := CreateFormPage(r, id)
		if err != nil {
			errorPage(w, "Error CreateFormPage", err, 500)
			return
		}

		fs, err := CreateFormFile(r, datastore.FileTypePageImage)
		if err != nil {
			errorPage(w, "Error CreateFormFile", err, 500)
			return
		}

		ps.ID = id
		if fs != nil {
			draftID := datastore.CreateDraftPageImageID(id)
			fs.ID = draftID
		}

		ps.Page = p
		ps.PageData = pd
		ps.FileSet = fs

		ctx := r.Context()

		err = datastore.PutPage(ctx, &ps)
		if err != nil {
			errorPage(w, "Error Put Page", err, 500)
			return
		}
	}

	vars := mux.Vars(r)
	id := vars["key"]

	//ページ検索
	page, err := datastore.SelectPage(ctx, id, -1)
	if err != nil {
		errorPage(w, "Error Select Page", err, 500)
		return
	}

	if page == nil {
		if id == datastore.ErrorPageID {
			page = &datastore.Page{}
			page.Deleted = true
			page.Parent = ""
			page.LoadKey(datastore.CreatePageKey(id))
		} else {
			//TODO ありえないけどどうしよう
		}
	}

	view(w, r, page)
}

func view(w http.ResponseWriter, r *http.Request, page *datastore.Page) {

	var err error
	var pageData *datastore.PageData

	publish := false

	ctx := r.Context()

	//全件検索
	templates, _, err := datastore.SelectTemplates(ctx, "all", datastore.NoLimitCursor)
	if err != nil {
		errorPage(w, "Error Select Template", err, 500)
		return
	}

	var children []datastore.Page
	siteTemplateName := "Select Site Template..."
	pageTemplateName := "Select Page Template..."

	id := page.Key.Name

	pageData, err = datastore.SelectPageData(ctx, id)
	if err != nil {
		errorPage(w, "Error Select PageData", err, 500)
		return
	}

	if pageData == nil {
		pageData = &datastore.PageData{}
		pageData.LoadKey(datastore.CreatePageDataKey(id))
	}

	//全件でOK
	children, _, err = datastore.SelectChildPages(ctx, id, datastore.NoLimitCursor, 0, true)
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
		if parent == "" || parent == datastore.ErrorPageID {
			break
		}
		parentPage, err := datastore.SelectPage(ctx, parent, -1)
		if err != nil {
			break
		}
		wk = append(wk, *parentPage)
		parent = parentPage.Parent
	}

	exist := datastore.ExistFile(ctx, id)

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
		ExistFile        bool
		Publish          bool
		SiteTemplateName string
		PageTemplateName string
	}{page, pageData, children, breadcrumbs, templates, exist, publish, siteTemplateName, pageTemplateName}

	viewManage(w, "page/edit.tmpl", dto)
}

func deletePageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	ctx := r.Context()

	err := datastore.RemovePage(ctx, id)
	if err != nil {
		errorPage(w, "Error Delete Page", err, 500)
		return
	}
	http.Redirect(w, r, "/manage/page/", 302)
}

func changePublicPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	ctx := r.Context()

	err := logic.PutHTML(ctx, id)
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

	err := datastore.RemoveHTML(ctx, id)
	if err != nil {
		errorPage(w, "Error Private HTML", err, 500)
		return
	}
	http.Redirect(w, r, "/manage/page/"+id, 302)
}

func toolPageHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]

	ctx := r.Context()

	children, _, err := datastore.SelectChildPages(ctx, id, datastore.NoLimitCursor, 0, true)
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

	ctx := r.Context()

	err := datastore.PutPageSequence(ctx, idCsv, enablesCsv, versionsCsv)
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

	tree, err := datastore.CreatePagesTree(r.Context())
	if err != nil {
		errorPage(w, "Error Page Tree", err, 500)
		return
	}

	dto := struct {
		Tree *datastore.Tree
	}{tree}

	viewManage(w, "page/tree.tmpl", dto)
}
