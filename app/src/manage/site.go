package manage

import(
	"datastore"

	"net/http"
	"html/template"
	"time"
)

type URL struct {
	URL string
	LastModified string
	Priority string
	Change string
	Image string
	Caption string
}

//setting画面
func (h Handler) ViewSetting(w http.ResponseWriter, r *http.Request) {
	site := datastore.GetSite(r)
	dto := struct {
		Site *datastore.Site
	} {site}
	h.parse(w, TEMPLATE_DIR + "site/edit.tmpl", dto)
}

//settingの更新
func (h Handler) EditSetting(w http.ResponseWriter, r *http.Request) {
	err := datastore.PutSite(r)
	if err != nil {
		h.errorPage(w,"Datastore site put Error",err.Error(),500)
		return
	}
}

func (h Handler) DownloadSitemap(w http.ResponseWriter, r *http.Request) {

	//URLを解析
	root := "https://www.hagoromo-shizuoka.com/"

	//Page全体でアクセス
	pages,err := datastore.SelectPages(r)
	if err != nil {
		h.errorPage(w,"Datastore select pages Error",err.Error(),500)
		return
	}

	urls := make([]URL,len(pages))
	//Page数回繰り返す
	for _,page := range pages {

		url := URL{}

		url.URL = root + "page/" + page.Key.StringID()
		url.LastModified = page.UpdatedAt.Format(time.RFC3339)
		url.Change = "weekly"
		url.Priority = "0.8"
		url.Image = root + "file/" + page.Key.StringID()
		url.Caption = page.Description
	}

	dto := struct {
		Pages []URL
	}{urls}

	//Topと同じだった場合
	tmpl, err := template.ParseFiles("templates/sitemap.tmpl")
	if err != nil {
		h.errorPage(w,"Sitemap template parse Error",err.Error(),500)
		return
	}

	err = tmpl.Execute(w, dto)
	if err != nil {
		h.errorPage(w,"Sitemap template execute Error",err.Error(),500)
		return
	}
}

