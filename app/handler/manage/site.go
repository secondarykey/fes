package manage

import (
	"app/datastore"

	"net/http"
)

//setting画面
func (h Handler) ViewSetting(w http.ResponseWriter, r *http.Request) {

	managers := ""
	site, err := datastore.SelectSite(r, -1)
	if err != nil {
		if err == datastore.SiteNotFoundError {
			site = &datastore.Site{
				Name:        "サイトの名前",
				Description: "サイトの説明",
			}
		} else {
			h.errorPage(w, "Site select error", err.Error(), 500)
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

	h.parse(w, TEMPLATE_DIR+"site/edit.tmpl", dto)
}

//settingの更新
func (h Handler) EditSetting(w http.ResponseWriter, r *http.Request) {
	err := datastore.PutSite(r)
	if err != nil {
		h.errorPage(w, "Datastore site put Error", err.Error(), 500)
		return
	}
	h.ViewSetting(w, r)
}

func (h Handler) DownloadSitemap(w http.ResponseWriter, r *http.Request) {

	err := datastore.GenerateSitemap(w, r)
	if err != nil {
		h.errorPage(w, "sitemap Error", err.Error(), 500)
		return
	}
}
