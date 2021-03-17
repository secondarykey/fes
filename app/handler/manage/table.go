package manage

import (
	"net/http"
)

func viewTableHandler(w http.ResponseWriter, r *http.Request) {

	dto := struct {
		Type string
	}{""}

	viewManage(w, "table/view.tmpl", dto)
}
