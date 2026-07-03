package internal

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:embed _assets/manage
var embManageSPA embed.FS
var manageSPAFs fs.FS

func init() {
	var err error
	manageSPAFs, err = fs.Sub(embManageSPA, "_assets/manage")
	if err != nil {
		log.Printf("manage-spa init error: %+v", err)
	}
}

func RegisterManageSPAStatic(mux *http.ServeMux) error {
	mux.Handle("/manage/assets/", http.StripPrefix("/manage/", http.FileServer(http.FS(manageSPAFs))))
	return nil
}

// ServeManageSPA はファイルが存在すればそのまま返し、なければ index.html を返す (SPA ルーティング対応)
func ServeManageSPA(w http.ResponseWriter, r *http.Request) {
	filePath := strings.TrimPrefix(r.URL.Path, "/manage")
	if filePath == "" || filePath == "/" {
		http.ServeFileFS(w, r, manageSPAFs, "index.html")
		return
	}
	name := strings.TrimLeft(filePath, "/")
	if _, err := fs.Stat(manageSPAFs, name); err == nil {
		http.ServeFileFS(w, r, manageSPAFs, name)
		return
	}
	http.ServeFileFS(w, r, manageSPAFs, "index.html")
}
