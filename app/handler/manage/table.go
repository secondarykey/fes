package manage

import (
	"net/http"
)

func (h Handler) TableView(w http.ResponseWriter, r *http.Request) {

	dto := struct {
		Type string
	} {""}

	h.parse(w,TEMPLATE_DIR + "table/view.tmpl",dto)
}

