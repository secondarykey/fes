package internal

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:embed _assets/manage-v2
var embManageV2 embed.FS
var manageV2Fs fs.FS

func init() {
	var err error
	manageV2Fs, err = fs.Sub(embManageV2, "_assets/manage-v2")
	if err != nil {
		log.Printf("manage-v2 init error: %+v", err)
	}
}

func RegisterManageV2Static() error {
	// /manage/v2/assets/ はセッションチェックなしで配信
	http.Handle("/manage/v2/assets/", http.StripPrefix("/manage/v2/", http.FileServer(http.FS(manageV2Fs))))
	return nil
}

// ServeManageV2SPA はファイルが存在すればそのまま返し、なければ index.html を返す (SPA ルーティング対応)
func ServeManageV2SPA(w http.ResponseWriter, r *http.Request) {
	filePath := strings.TrimPrefix(r.URL.Path, "/manage/v2")
	if filePath == "" || filePath == "/" {
		http.ServeFileFS(w, r, manageV2Fs, "index.html")
		return
	}
	name := strings.TrimLeft(filePath, "/")
	if _, err := fs.Stat(manageV2Fs, name); err == nil {
		http.ServeFileFS(w, r, manageV2Fs, name)
		return
	}
	http.ServeFileFS(w, r, manageV2Fs, "index.html")
}
