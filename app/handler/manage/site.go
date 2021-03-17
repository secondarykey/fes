package manage

import (
	"app/datastore"

	"net/http"
)

//setting画面
func viewSiteHandler(w http.ResponseWriter, r *http.Request) {

	managers := ""
	site, err := datastore.SelectSite(r, -1)
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
	err := datastore.PutSite(r)
	if err != nil {
		errorPage(w, "Datastore site put Error", err, 500)
		return
	}
	//TODO redirect???
	viewSiteHandler(w, r)
}

func downloadSitemapHandler(w http.ResponseWriter, r *http.Request) {

	err := datastore.GenerateSitemap(w, r)
	if err != nil {
		errorPage(w, "sitemap Error", err, 500)
		return
	}
}
