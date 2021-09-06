package handler

import (
	"app/datastore"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func fileHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "public, max-age=3600")
	//ファイルを検索
	vars := mux.Vars(r)
	id := vars["key"]

	//表示
	fileData, err := datastore.GetFileData(r.Context(), id)
	if err != nil {
		errorPage(w, r, "Datastore:FileData Search Error", err, 500)
		return
	}

	if fileData == nil {
		errorPage(w, r, "Datastore:Not Found FileData Error", fmt.Errorf("指定したIDのデータが存在しません。%s", id), 404)
		return
	}

	w.Header().Set("Content-Type", fileData.Mime)
	_, err = w.Write(fileData.Content)
	if err != nil {
		errorPage(w, r, "Writing FileData Error", err, 500)
		return
	}
	return
}

func fileDateCacheHandler(w http.ResponseWriter, r *http.Request) {
	// 60 * 60 * 24 = 86400
	// * 10 = 864000
	w.Header().Set("Cache-Control", "public, max-age=864000")
	fileHandler(w, r)
}

func fileCacheHandler(w http.ResponseWriter, r *http.Request) {
	// 60 * 60 * 3  = 10800
	// 60 * 60 * 6  = 21600
	w.Header().Set("Cache-Control", "public, max-age=21600")
	fileHandler(w, r)
}
