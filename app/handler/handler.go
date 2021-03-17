package handler

import (
	"app/datastore"
	. "app/handler/internal"
	"fmt"
	"log"

	"net/http"

	"github.com/gorilla/mux"
)

func Register() error {

	//外部アクセス
	r := mux.NewRouter()

	r.HandleFunc("/page/{key}", pageHandler).Methods("GET")
	r.HandleFunc("/file/{key}", fileHandler).Methods("GET")
	r.HandleFunc("/file/{date}/{key}", fileDateCacheHandler).Methods("GET")
	r.HandleFunc("/file_cache/{key}", fileCacheHandler).Methods("GET")

	r.HandleFunc("/login", loginHandler).Methods("GET")
	r.HandleFunc("/session", sessionHandler).Methods("POST")
	r.HandleFunc("/sitemap/", sitemap).Methods("GET")
	r.HandleFunc("/", indexHandler).Methods("GET")

	http.Handle("/", r)

	return nil
}

func errorPage(w http.ResponseWriter, t string, e error, num int) {

	msg := fmt.Sprintf("%+v", e)

	log.Println(msg)

	dto := struct {
		Title   string
		Message string
		No      int
	}{t, msg, num}

	View(w, dto, "error.tmpl")
}

func sitemap(w http.ResponseWriter, r *http.Request) {
	// 60 * 60 * 24
	w.Header().Set("Cache-Control", "public, max-age=86400")
	err := datastore.GenerateSitemap(w, r)
	if err != nil {
		errorPage(w, "Generate sitemap error", err, 500)
	}
}
