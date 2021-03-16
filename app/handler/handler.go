package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func Register() error {

	//外部アクセス
	r := mux.NewRouter()
	pub := Public{r}

	r.HandleFunc("/page/{key}", pub.pageHandler).Methods("GET")
	r.HandleFunc("/file/{key}", pub.fileHandler).Methods("GET")
	r.HandleFunc("/file/{date}/{key}", pub.fileDateCacheHandler).Methods("GET")
	r.HandleFunc("/file_cache/{key}", pub.fileCacheHandler).Methods("GET")

	r.HandleFunc("/login", pub.loginHandler).Methods("GET")
	r.HandleFunc("/session", pub.sessionHandler).Methods("POST")
	r.HandleFunc("/sitemap/", pub.sitemap).Methods("GET")
	r.HandleFunc("/", pub.topHandler).Methods("GET")

	http.Handle("/", pub)
	return nil
}
