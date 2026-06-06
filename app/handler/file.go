package handler

import (
	"app/datastore"
	"app/handler/manage"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func fileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	file, err := dao.SelectFile(ctx, id)
	if err != nil {
		errorPage(w, r, "Datastore error", err, 500)
		return
	}
	if file == nil {
		errorPage(w, r, "Datastore:Not Found FileData Error", fmt.Errorf("指定したIDのデータが存在しません。%s", id), 404)
		return
	}

	setCacheHeaders(w, file.UpdatedAt)
	if checkNotModified(r, file.UpdatedAt) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if err := manage.FileViewHandler(w, r); err != nil {
		errorPage(w, r, "Datastore:Not Found FileData Error", fmt.Errorf("指定したIDのデータが存在しません。%s", id), 404)
	}
}

// Deprecated: タイムスタンプ付き URL (/file/{date}/{key}) は既存テンプレートとの互換性のために残しています。
// 新規テンプレートでは /file/{key} を使用してください。
func fileDateCacheHandler(w http.ResponseWriter, r *http.Request) {
	fileHandler(w, r)
}

func fileCacheHandler(w http.ResponseWriter, r *http.Request) {
	fileHandler(w, r)
}
