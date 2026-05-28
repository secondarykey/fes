package manage

import (
	"app/datastore"
	"app/handler/manage/form"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func registerFileAPI(r *mux.Router) {
	r.HandleFunc("/file/", apiListFiles).Methods("GET")
	r.HandleFunc("/file/", apiUploadFile).Methods("POST")
	r.HandleFunc("/file/", apiDeleteFiles).Methods("DELETE")
	r.HandleFunc("/file/{key}", apiGetFile).Methods("GET")
	r.HandleFunc("/file/{key}", apiDeleteFile).Methods("DELETE")
	r.HandleFunc("/file/{key}/rename", apiRenameFile).Methods("POST")
	r.HandleFunc("/file/{key}/replace", apiReplaceFile).Methods("POST")
}

type apiFileRes struct {
	ID        string    `json:"id"`
	Size      int64     `json:"size"`
	Type      int       `json:"type"`
	Deleted   bool      `json:"deleted"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toAPIFileRes(f *datastore.File) apiFileRes {
	id := ""
	if f.Key != nil {
		id = f.Key.Name
	}
	return apiFileRes{
		ID:        id,
		Size:      f.Size,
		Type:      f.Type,
		Deleted:   f.Deleted,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}
}

func apiListFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	typ := r.URL.Query().Get("type")
	cursor := r.URL.Query().Get("cursor")

	files, nextCursor, err := dao.SelectFiles(ctx, typ, cursor)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	res := make([]apiFileRes, len(files))
	for i := range files {
		res[i] = toAPIFileRes(&files[i])
	}

	apiJSON(w, map[string]interface{}{
		"files":      res,
		"nextCursor": nextCursor,
	})
}

func apiUploadFile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		apiError(w, "multipart parse error", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	fs := &datastore.FileSet{}
	if err := form.SetFile(r, fs, datastore.FileTypeData); err != nil {
		apiError(w, err.Error(), 400)
		return
	}
	if fs.File == nil {
		apiError(w, "ファイルが選択されていません", 400)
		return
	}

	if err := dao.SaveFile(ctx, fs); err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	saved, err := dao.SelectFile(ctx, fs.Name)
	if err != nil || saved == nil {
		apiJSON(w, map[string]string{"status": "uploaded"})
		return
	}
	apiJSON(w, toAPIFileRes(saved))
}

func apiDeleteFiles(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IDs) == 0 {
		apiError(w, "ids required", 400)
		return
	}
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.RemoveFiles(ctx, req.IDs); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"status": "deleted"})
}

func apiGetFile(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	f, err := dao.SelectFile(ctx, id)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	if f == nil {
		apiError(w, "not found", 404)
		return
	}
	apiJSON(w, toAPIFileRes(f))
}

func apiDeleteFile(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.RemoveFile(ctx, id); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"status": "deleted"})
}

func apiRenameFile(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		apiError(w, "name は必須です", 400)
		return
	}
	if req.Name == id {
		apiError(w, "同じ名前です", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	// 既存ファイルのデータを取得
	oldFile, err := dao.SelectFile(ctx, id)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	if oldFile == nil {
		apiError(w, "ファイルが見つかりません", 404)
		return
	}

	oldData, err := dao.GetFileData(ctx, id)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	if oldData == nil {
		apiError(w, "ファイルデータが見つかりません", 404)
		return
	}

	// 新しい名前で保存
	fs := &datastore.FileSet{
		Name:     req.Name,
		File:     oldFile,
		FileData: oldData,
	}
	if err := dao.SaveFile(ctx, fs); err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	// 旧データを削除
	if err := dao.RemoveFile(ctx, id); err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	saved, err := dao.SelectFile(ctx, req.Name)
	if err != nil || saved == nil {
		apiJSON(w, map[string]string{"status": "renamed"})
		return
	}
	apiJSON(w, toAPIFileRes(saved))
}

func apiReplaceFile(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		apiError(w, "multipart parse error", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	// 既存ファイルの存在確認
	if !dao.ExistFile(ctx, id) {
		apiError(w, "ファイルが見つかりません", 404)
		return
	}

	// アップロードファイルを読み込み（名前は既存のIDを使う）
	fs := &datastore.FileSet{}
	if err := form.SetFile(r, fs, datastore.FileTypeData); err != nil {
		apiError(w, err.Error(), 400)
		return
	}
	if fs.File == nil {
		apiError(w, "ファイルが選択されていません", 400)
		return
	}

	// 既存のID（ファイル名）で上書き保存
	fs.ID = id
	fs.Name = id
	if err := dao.SaveFile(ctx, fs); err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	saved, err := dao.SelectFile(ctx, id)
	if err != nil || saved == nil {
		apiJSON(w, map[string]string{"status": "replaced"})
		return
	}
	apiJSON(w, toAPIFileRes(saved))
}
