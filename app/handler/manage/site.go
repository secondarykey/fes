package manage

import (
	"app/datastore"
	. "app/handler/internal"
	"app/handler/manage/form"

	"errors"
	"fmt"

	"net/http"
)

//setting画面
func viewSiteHandler(w http.ResponseWriter, r *http.Request) {

	dao := datastore.NewDao()
	defer dao.Close()

	managers := ""
	ctx := r.Context()
	site, err := dao.SelectSite(ctx, -1)
	if err != nil {
		if err == datastore.SiteNotFoundError {
			site = &datastore.Site{
				Name:        "サイトの名前",
				Description: "サイトの説明",
			}
		} else {
			errorPage(w, "Site select error", err, 500)
			return
		}
	} else {

		for _, mail := range site.Managers {
			if managers != "" {
				managers += ","
			}
			managers += mail
		}
	}

	dto := struct {
		Site     *datastore.Site
		Managers string
	}{site, managers}

	viewManage(w, "site/edit.tmpl", dto)
}

//settingの更新
func editSiteHandler(w http.ResponseWriter, r *http.Request) {

	dao := datastore.NewDao()
	defer dao.Close()

	ctx := r.Context()

	site, err := dao.SelectSite(ctx, -1)
	if err != nil {
		if !errors.Is(err, datastore.SiteNotFoundError) {
			errorPage(w, "dao SelectSite() Error", err, 500)
			return
		}
	}

	if site == nil {
		site = new(datastore.Site)
	}

	err = form.SetSite(r, site)
	if err != nil {
		errorPage(w, "CreateFormSite() Error", err, 500)
		return
	}

	err = dao.PutSite(ctx, site)
	if err != nil {
		errorPage(w, "Datastore site put Error", err, 500)
		return
	}

	_, err = dao.GetTrashPage(ctx)
	if err != nil {
		errorPage(w, "Datastore trash Page Error", err, 500)
		return
	}

	viewSiteHandler(w, r)
}

func cleanSiteHandler(w http.ResponseWriter, r *http.Request) {

	//全HTMLを検索
	dao := datastore.NewDao()
	defer dao.Close()

	ctx := r.Context()
	ids, err := dao.GetHTMLs(ctx)
	if err != nil {
		errorPage(w, "HTML Page Error", err, 500)
		return
	}

	fmt.Println("HTML", len(ids))

	pages, err := dao.SelectPages(ctx, ids...)
	newPage := make([]datastore.Page, 0, len(pages))

	if err != nil {
		for _, p := range pages {
			if p.Key != nil {
				newPage = append(newPage, p)
			}
		}
	}

	//HTMLがあるがページがない HTML削除
	//浮いているページ全部のページとページツリーの突合

	for _, id := range ids {
		nf := true
		for _, p := range newPage {
			work := p.Key.Name
			if id == work {
				nf = false
				break
			}
		}

		if nf {
			fmt.Println(id)
		}
	}

	dto := struct {
	}{}

	viewManage(w, "site/clean.tmpl", dto)
}

func downloadSitemapHandler(w http.ResponseWriter, r *http.Request) {

	scheme := r.URL.Scheme
	if scheme == "" {
		scheme = "https"
	}
	root := fmt.Sprintf("%s://%s/", scheme, r.Host)

	err := GenerateSitemap(r.Context(), root, w)
	if err != nil {
		errorPage(w, "Error GenerateSitemap()", err, 500)
		return
	}
}
