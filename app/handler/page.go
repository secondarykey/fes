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
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	site, err := dao.SelectSite(ctx, -1)
	if err != nil {
		errorPage(w, r, "Not Found", fmt.Errorf("サイトにトップページが指定されていません。"), 404)
		return
	}
	pageView(w, r, site.Root)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]
	pageView(w, r, id)
}

func pageView(w http.ResponseWriter, r *http.Request, id string) {
	val := r.URL.Query()
	page := val.Get("page")
	if page != "" {
		id += "?page=" + page
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	html, err := dao.GetHTML(ctx, id)
	if err != nil {
		errorPage(w, r, "datastore.GetHTML() error", err, 500)
		return
	}
	if html == nil {
		errorPage(w, r, "page not found", xerrors.Errorf("Page NotFound: %w", fmt.Errorf("%s", id)), 404)
		return
	}

	setCacheHeaders(w, html.UpdatedAt)
	if checkNotModified(r, html.UpdatedAt) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(html.Content); err != nil {
		log.Println("Write Error", err)
	}
}
