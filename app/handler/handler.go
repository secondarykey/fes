package handler

import (
	. "app/handler/internal"
	"fmt"
	"log"

	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
)

func Register() error {

	err := RegisterArchive("2020")
	if err != nil {
		return xerrors.Errorf("RegisterArchive() error: %w", err)
	}

	//外部アクセス
	r := mux.NewRouter()

	err = RegisterStatic()
	if err != nil {
		return xerrors.Errorf("RegisterStatic() error: %w", err)
	}

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
	scheme := r.URL.Scheme
	if scheme == "" {
		scheme = "https"
	}
	root := fmt.Sprintf("%s://%s/", scheme, r.Host)

	// 60 * 60 * 24
	w.Header().Set("Cache-Control", "public, max-age=86400")
	err := GenerateSitemap(r.Context(), root, w)
	if err != nil {
		errorPage(w, "Generate sitemap error", err, 500)
	}
}
