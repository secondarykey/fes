package manage

import (
	"app/datastore"
	"app/handler/manage/form"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func registerV2FileAPI(r *mux.Router) {
	r.HandleFunc("/file/", v2ListFiles).Methods("GET")
	r.HandleFunc("/file/", v2UploadFile).Methods("POST")
	r.HandleFunc("/file/", v2DeleteFiles).Methods("DELETE")
	r.HandleFunc("/file/{key}", v2DeleteFile).Methods("DELETE")
}

type v2FileRes struct {
	ID        string    `json:"id"`
	Size      int64     `json:"size"`
	Type      int       `json:"type"`
	Deleted   bool      `json:"deleted"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toV2FileRes(f *datastore.File) v2FileRes {
	id := ""
	if f.Key != nil {
		id = f.Key.Name
	}
	return v2FileRes{
		ID:        id,
		Size:      f.Size,
		Type:      f.Type,
		Deleted:   f.Deleted,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}
}

func v2ListFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	typ := r.URL.Query().Get("type")
	cursor := r.URL.Query().Get("cursor")

	files, nextCursor, err := dao.SelectFiles(ctx, typ, cursor)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	res := make([]v2FileRes, len(files))
	for i := range files {
		res[i] = toV2FileRes(&files[i])
	}

	v2JSON(w, map[string]interface{}{
		"files":      res,
		"nextCursor": nextCursor,
	})
}

func v2UploadFile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		v2Error(w, "multipart parse error", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	fs := &datastore.FileSet{}
	if err := form.SetFile(r, fs, datastore.FileTypeData); err != nil {
		v2Error(w, err.Error(), 400)
		return
	}
	if fs.File == nil {
		v2Error(w, "ファイルが選択されていません", 400)
		return
	}

	if err := dao.SaveFile(ctx, fs); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	saved, err := dao.SelectFile(ctx, fs.Name)
	if err != nil || saved == nil {
		v2JSON(w, map[string]string{"status": "uploaded"})
		return
	}
	v2JSON(w, toV2FileRes(saved))
}

func v2DeleteFiles(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IDs) == 0 {
		v2Error(w, "ids required", 400)
		return
	}
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.RemoveFiles(ctx, req.IDs); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]string{"status": "deleted"})
}

func v2DeleteFile(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.RemoveFile(ctx, id); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]string{"status": "deleted"})
}
