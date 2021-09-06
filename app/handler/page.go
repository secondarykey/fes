package handler

import (
	"app/datastore"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=3600")
	ctx := r.Context()
	site, err := datastore.SelectSite(ctx, -1)
	if err != nil {
		errorPage(w, r, "Not Found", fmt.Errorf("サイトにトップページが指定されていません。"), 404)
		return
	}
	pageView(w, r, site.Root)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=3600")
	vars := mux.Vars(r)
	id := vars["key"]
	pageView(w, r, id)
}

func pageView(w http.ResponseWriter, r *http.Request, id string) {

	//ページを取得してIDを作成
	val := r.URL.Query()
	page := val.Get("page")
	if page != "" {
		id += "?page=" + page
	}

	html, err := datastore.GetHTML(r.Context(), id)
	if err != nil {
		errorPage(w, r, "datastore.GetHTML() error", err, 500)
		return
	}
	if html == nil {
		errorPage(w, r, "page not found", xerrors.Errorf("Page NotFound: %w", fmt.Errorf("%s", id)), 404)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	_, err = w.Write(html.Content)
	if err != nil {
		log.Println("Write Error", err)
	}
}
