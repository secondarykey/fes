package internal

import (
	"net/http"
)

func RegisterManageStatic() error {

	fs := http.StripPrefix("/manage/", http.FileServer(GrantFS(statikFS, "/static/manage")))

	http.Handle("/manage/js/", fs)
	http.Handle("/manage/css/", fs)

	return nil
}
