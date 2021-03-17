package handler

import (
	"app/datastore"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=3600")
	site, err := datastore.SelectSite(r, -1)
	if err != nil {
		errorPage(w, "Not Found", fmt.Errorf("サイトにトップページが指定されていません。"), 404)
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

	w.Header().Set("Cache-Control", "public, max-age=3600")
	//ページを取得してIDを作成
	val := r.URL.Query()
	page := val.Get("page")
	if page != "" {
		id += "?page=" + page
	}

	html, err := datastore.GetHTML(r.Context(), id)
	if err != nil {
		errorPage(w, "error get html", err, 500)
		return
	}
	if html == nil {
		errorPage(w, "page not found", fmt.Errorf("ページが存在しません。[%s]", id), 404)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	_, err = w.Write(html.Content)
	if err != nil {
		log.Println("Write Error", err)
	}
}
