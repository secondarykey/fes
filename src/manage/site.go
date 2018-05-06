package manage

import(
	"datastore"

	"net/http"
)

func (h Handler) ViewSetting(w http.ResponseWriter, r *http.Request) {
	site := datastore.GetSite(r)
	dto := struct {
		Site *datastore.Site
	} {site}
	h.parse(w, TEMPLATE_DIR + "site/edit.tmpl", dto)
}

func (h Handler) EditSetting(w http.ResponseWriter, r *http.Request) {
	err := datastore.PutSite(r)
	if err != nil {
		h.errorPage(w,"Datastore site put Error",err.Error(),500)
	}
}


