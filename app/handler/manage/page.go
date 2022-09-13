package manage

import (
	"app/datastore"
	"app/handler/manage/form"

	"net/http"

	"github.com/gorilla/mux"
)

func viewRootPageHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	page, err := dao.SelectRootPage(ctx)
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

	page := datastore.Page{}

	page.Parent = parent
	//削除は一覧に表示されない仕様に変更されました
	page.Deleted = false

	page.LoadKey(datastore.CreatePageKey())

	view(w, r, &page)
}

func viewPageHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if POST(r) {

		vars := mux.Vars(r)
		id := vars["key"]

		p, err := dao.SelectPage(ctx, id, -1)
		if err != nil {
			errorPage(w, "Error CreateFormPage", err, 500)
			return
		}

		ps := new(datastore.PageSet)
		ps.Page = p
		if p == nil {
			ps.Page = &datastore.Page{}
		}
		ps.PageData = new(datastore.PageData)

		err = form.SetPage(r, ps, id)
		if err != nil {
			errorPage(w, "Error CreateFormPage", err, 500)
			return
		}

		fs := new(datastore.FileSet)

		err = form.SetFile(r, fs, datastore.FileTypePageImage)
		if err != nil {
			errorPage(w, "Error CreateFormFile", err, 500)
			return
		}

		ps.ID = id
		if fs != nil {
			draftID := datastore.CreateDraftPageImageID(id)
			fs.ID = draftID
		}

		ps.FileSet = fs
		ctx := r.Context()

		err = dao.PutPage(ctx, ps)
		if err != nil {
			errorPage(w, "Error Put Page", err, 500)
			return
		}
	}

	vars := mux.Vars(r)
	id := vars["key"]

	//ページ検索
	page, err := dao.SelectPage(ctx, id, -1)
	if err != nil {
		errorPage(w, "Error Select Page", err, 500)
		return
	}

	if page == nil {
		if id == datastore.ErrorPageID {
			page = &datastore.Page{}
			page.Deleted = true
			page.Parent = ""
			page.LoadKey(datastore.CreatePageKey())
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
	dao := datastore.NewDao()
	defer dao.Close()

	//全件検索
	templates, _, err := dao.SelectTemplates(ctx, "all", datastore.NoLimitCursor)
	if err != nil {
		errorPage(w, "Error Select Template", err, 500)
		return
	}

	var children []datastore.Page
	siteTemplateName := "Select Site Template..."
	pageTemplateName := "Select Page Template..."

	id := page.Key.Name

	pageData, err = dao.SelectPageData(ctx, id)
	if err != nil {
		errorPage(w, "Error Select PageData", err, 500)
		return
	}

	if pageData == nil {
		pageData = &datastore.PageData{}
		pageData.LoadKey(datastore.GetPageDataKey(id))
	}

	//全件でOK
	children, _, err = dao.SelectChildrenPage(ctx, id, datastore.NoLimitCursor, 0, true)
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
		parentPage, err := dao.SelectPage(ctx, parent, -1)
		if err != nil {
			break
		}
		wk = append(wk, *parentPage)
		parent = parentPage.Parent
	}

	exist := dao.ExistFile(ctx, id)
	existDraft := dao.ExistFile(ctx, datastore.CreateDraftPageImageID(id))

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
		ExistFile        bool
		ExistDraftFile   bool
		Publish          bool
	}{page, pageData, children, breadcrumbs,
		templates, siteTemplateName, pageTemplateName,
		exist, existDraft, publish}

	viewManage(w, "page/edit.tmpl", dto)
}

func deletePageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	err := dao.RemovePage(ctx, id)
	if err != nil {
		errorPage(w, "Error Delete Page", err, 500)
		return
	}
	http.Redirect(w, r, "/manage/page/", 302)
}
