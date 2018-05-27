package manage

import(
	"datastore"

	"net/http"
)

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

//初回アクセスにおける設定
func (h Handler) FirstSetting(r *http.Request) {

	//テンプレートを登録

    //サイトを登録

    //最初のページを設定

}
