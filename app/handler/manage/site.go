package manage

import (
	"app/datastore"
	. "app/handler/internal"
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

	site, err := CreateFormSite(r)
	if err != nil {
		errorPage(w, "CreateFormSite() Error", err, 500)
		return
	}

	ctx := r.Context()
	err = dao.PutSite(ctx, site)
	if err != nil {
		errorPage(w, "Datastore site put Error", err, 500)
		return
	}
	//TODO redirect???
	viewSiteHandler(w, r)
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
