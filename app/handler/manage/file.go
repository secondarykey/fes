package manage

import (
	"app/datastore"

	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
)

func fileViewHandler(w http.ResponseWriter, r *http.Request) {
	err := FileViewHandler(w, r)
	if err != nil {
		errorPage(w, "Error file View", err, 404)
	}
}

func FileViewHandler(w http.ResponseWriter, r *http.Request) error {
	//ファイルを検索
	vars := mux.Vars(r)
	id := vars["key"]

	dao := datastore.NewDao()
	defer dao.Close()

	//表示
	fileData, err := dao.GetFileData(r.Context(), id)
	if err != nil {
		return xerrors.Errorf("GetFileData() error: %w", err)
	}

	if fileData == nil {
		return fmt.Errorf("FileData is nil: %s", id)
	}

	w.Header().Set("Content-Type", fileData.Mime)
	_, err = w.Write(fileData.Content)
	if err != nil {
		return xerrors.Errorf("Writer Write() error: %w", err)
	}
	return nil
}
