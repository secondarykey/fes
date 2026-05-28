package manage

import (
	"app/datastore"
	"app/logic"
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func registerToolAPI(r *mux.Router) {
	r.HandleFunc("/tool/gc", apiGC).Methods("POST")
	r.HandleFunc("/tool/clean", apiSiteClean).Methods("POST")
	r.HandleFunc("/tool/page-list", apiPageList).Methods("GET")
	r.HandleFunc("/tool/backup", apiBackup).Methods("GET")
	r.HandleFunc("/tool/restore", apiRestore).Methods("POST")
}

func apiGC(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	if err := logic.GC(&buf); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"result": buf.String()})
}

func apiSiteClean(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	result := struct {
		DeletedHTMLs   []string `json:"deletedHtmls"`
		DeletedImages  []string `json:"deletedImages"`
		Errors         []string `json:"errors"`
	}{}

	// 孤立HTML削除
	htmlIDs, err := dao.GetHTMLs(ctx)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	for _, id := range htmlIDs {
		page, err := dao.SelectPage(ctx, id, -1)
		if err != nil || page == nil {
			if err := dao.RemoveOrphanHTML(ctx, id); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("HTML[%s]: %v", id, err))
			} else {
				result.DeletedHTMLs = append(result.DeletedHTMLs, id)
			}
		}
	}

	// 孤立ページ画像削除（公開・下書き両方）
	files, err := dao.GetAllFiles(ctx)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	for _, f := range files {
		if f.Type != datastore.FileTypePageImage {
			continue
		}
		fileID := f.Key.Name
		// 下書き画像のIDからページIDを取り出す
		pageID := strings.TrimSuffix(fileID, "-"+datastore.DraftPageImageIDSuffix)
		page, err := dao.SelectPage(ctx, pageID, -1)
		if err != nil || page == nil {
			if err := dao.RemoveFile(ctx, fileID); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("File[%s]: %v", fileID, err))
			} else {
				result.DeletedImages = append(result.DeletedImages, fileID)
			}
		}
	}

	apiJSON(w, result)
}

func apiPageList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	parentID := r.URL.Query().Get("parent")
	if parentID == "" {
		root, err := dao.SelectRootPage(ctx)
		if err != nil || root == nil {
			apiError(w, "root page not found", 500)
			return
		}
		parentID = root.Key.Name
	}

	var buf bytes.Buffer
	if err := appendPageListItems(ctx, dao, parentID, &buf); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"html": buf.String()})
}

func apiBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	data, err := dao.CreateBackupData(ctx)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	now := time.Now()
	filename := "fes-backup-" + now.Format("20060102150405") + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)

	zw := zip.NewWriter(w)
	defer zw.Close()

	for kind, elm := range data {
		for key, byt := range elm {
			esp := strings.ReplaceAll(key, "?", "_QUESTION_")
			esp = strings.ReplaceAll(esp, "=", "_EQUAL_")
			fw, err := zw.Create(kind + "/" + esp)
			if err != nil {
				return
			}
			fw.Write(byt)
		}
	}
}

func apiRestore(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		apiError(w, "ファイルの読み込みに失敗しました: "+err.Error(), 400)
		return
	}
	defer file.Close()

	reader, err := zip.NewReader(file, header.Size)
	if err != nil {
		apiError(w, "ZIP の解析に失敗しました: "+err.Error(), 400)
		return
	}

	backup, err := createGob(reader)
	if err != nil {
		apiError(w, "データの変換に失敗しました: "+err.Error(), 500)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.PutBackupData(ctx, backup); err != nil {
		apiError(w, "リストアに失敗しました: "+err.Error(), 500)
		return
	}

	apiJSON(w, map[string]string{"status": "restored"})
}

func appendPageListItems(ctx context.Context, dao *datastore.Dao, parentID string, buf *bytes.Buffer) error {
	children, _, err := dao.SelectChildrenPage(ctx, parentID, datastore.NoLimitCursor, 0, false)
	if err != nil {
		return err
	}
	for _, p := range children {
		id := p.Key.Name
		fmt.Fprintf(buf, "<li><a href=\"{{.Dir}}%s\">%s</a></li>\n", id, p.Name)
		if err := appendPageListItems(ctx, dao, id, buf); err != nil {
			return err
		}
	}
	return nil
}
