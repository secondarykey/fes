package internal

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed _assets/static
var embStatic embed.FS
var staticFs fs.FS

func init() {
	var err error
	staticFs, err = fs.Sub(embStatic, "_assets/static/manage")
	if err != nil {
		log.Printf("%+v", err)
	}
}

func RegisterManageStatic() error {

	fs := http.StripPrefix("/manage/", http.FileServer(http.FS(staticFs)))

	http.Handle("/manage/js/", fs)
	http.Handle("/manage/css/", fs)

	return nil
}
