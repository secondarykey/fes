package manage

import (
	"app/datastore"
	"app/logic"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// multipart の最大サイズ (32 MB)
const maxImageSize = 32 << 20

func registerV2PageAPI(r *mux.Router) {
	r.HandleFunc("/page/", v2GetRootPage).Methods("GET")
	// /page/new/{parentKey} は /page/{key} より先に登録する
	r.HandleFunc("/page/new/{parentKey}", v2NewPage).Methods("GET")
	r.HandleFunc("/page/children/{key}", v2GetChildren).Methods("GET")
	r.HandleFunc("/page/{key}/image", v2GetPageImage).Methods("GET")
	r.HandleFunc("/page/{key}/image", v2UploadPageImage).Methods("POST")
	r.HandleFunc("/page/{key}/image", v2DeletePageImage).Methods("DELETE")
	r.HandleFunc("/page/{key}/sequence", v2PageSequence).Methods("POST")
	r.HandleFunc("/page/{key}/sort", v2PageSort).Methods("POST")
	r.HandleFunc("/page/{key}/move", v2PageMove).Methods("POST")
	r.HandleFunc("/page/{key}", v2GetPage).Methods("GET")
	r.HandleFunc("/page/{key}", v2UpdatePage).Methods("POST")
	r.HandleFunc("/page/{key}", v2DeletePage).Methods("DELETE")
	r.HandleFunc("/page/", v2DeletePages).Methods("DELETE")
	r.HandleFunc("/html/publish/{key}", v2PublishPage).Methods("POST")
	r.HandleFunc("/html/unpublish/{key}", v2UnpublishPage).Methods("POST")
	r.HandleFunc("/html/publish-pages", v2PublishPages).Methods("POST")
}

// ── レスポンス型 ──────────────────────────────────────────

type v2PageRes struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Seq          int       `json:"seq"`
	Description  string    `json:"description"`
	Parent       string    `json:"parent"`
	SiteTemplate string    `json:"siteTemplate"`
	PageTemplate string    `json:"pageTemplate"`
	Paging       int       `json:"paging"`
	Deleted      bool      `json:"deleted"`
	Version      int       `json:"version"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	CanPublish   bool      `json:"canPublish"`
}

type v2TemplateRes struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"`
	Seq  int    `json:"seq"`
}

func toV2PageRes(p *datastore.Page) v2PageRes {
	if p == nil {
		return v2PageRes{}
	}
	id := ""
	if p.Key != nil {
		id = p.Key.Name
	}
	canPublish := !p.Deleted && p.UpdatedAt.Unix() > p.Republish.Unix()+1
	return v2PageRes{
		ID:           id,
		Name:         p.Name,
		Seq:          p.Seq,
		Description:  p.Description,
		Parent:       p.Parent,
		SiteTemplate: p.SiteTemplate,
		PageTemplate: p.PageTemplate,
		Paging:       p.Paging,
		Deleted:      p.Deleted,
		Version:      p.Version,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		CanPublish:   canPublish,
	}
}

func v2JSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func v2Error(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func v2Templates(r *http.Request, dao *datastore.Dao) ([]v2TemplateRes, error) {
	templates, _, err := dao.SelectTemplates(r.Context(), "all", datastore.NoLimitCursor)
	if err != nil {
		return nil, err
	}
	res := make([]v2TemplateRes, len(templates))
	for i, t := range templates {
		res[i] = v2TemplateRes{ID: t.Key.Name, Name: t.Name, Type: t.Type, Seq: t.Seq}
	}
	return res, nil
}

// ── ハンドラ ──────────────────────────────────────────────

func v2GetRootPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	page, err := dao.SelectRootPage(ctx)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	if page == nil {
		v2Error(w, "root page not found", 404)
		return
	}

	children, _, err := dao.SelectChildrenPage(ctx, page.Key.Name, datastore.NoLimitCursor, 0, true)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	childRes := make([]v2PageRes, len(children))
	for i := range children {
		childRes[i] = toV2PageRes(&children[i])
	}

	v2JSON(w, map[string]interface{}{"page": toV2PageRes(page), "children": childRes})
}

func v2GetChildren(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	children, _, err := dao.SelectChildrenPage(ctx, id, datastore.NoLimitCursor, 0, true)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	res := make([]v2PageRes, len(children))
	for i := range children {
		res[i] = toV2PageRes(&children[i])
	}
	v2JSON(w, res)
}

func v2NewPage(w http.ResponseWriter, r *http.Request) {
	parentKey := mux.Vars(r)["parentKey"]
	dao := datastore.NewDao()
	defer dao.Close()

	page := &datastore.Page{}
	page.Parent = parentKey
	page.Deleted = false
	page.LoadKey(datastore.CreatePageKey())

	tmplRes, err := v2Templates(r, dao)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	v2JSON(w, map[string]interface{}{
		"page":        toV2PageRes(page),
		"pageData":    "",
		"children":    []v2PageRes{},
		"breadcrumbs": []v2PageRes{},
		"templates":   tmplRes,
	})
}

func v2GetPage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	page, err := dao.SelectPage(ctx, id, -1)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	if page == nil {
		v2Error(w, "page not found", 404)
		return
	}

	pageData, err := dao.SelectPageData(ctx, id)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	content := ""
	if pageData != nil {
		content = string(pageData.Content)
	}

	children, _, err := dao.SelectChildrenPage(ctx, id, datastore.NoLimitCursor, 0, true)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	childRes := make([]v2PageRes, len(children))
	for i := range children {
		childRes[i] = toV2PageRes(&children[i])
	}

	// パンくずリスト
	var breadcrumbs []v2PageRes
	parent := page.Parent
	for parent != "" {
		pp, err := dao.SelectPage(ctx, parent, -1)
		if err != nil || pp == nil {
			break
		}
		breadcrumbs = append([]v2PageRes{toV2PageRes(pp)}, breadcrumbs...)
		parent = pp.Parent
	}

	tmplRes, err := v2Templates(r, dao)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	v2JSON(w, map[string]interface{}{
		"page":        toV2PageRes(page),
		"pageData":    content,
		"children":    childRes,
		"breadcrumbs": breadcrumbs,
		"templates":   tmplRes,
	})
}

type v2PageUpdateReq struct {
	Name         string `json:"name"`
	ParentID     string `json:"parentId"`
	Seq          int    `json:"seq"`
	Description  string `json:"description"`
	SiteTemplate string `json:"siteTemplate"`
	PageTemplate string `json:"pageTemplate"`
	Paging       int    `json:"paging"`
	Published    bool   `json:"published"`
	Content      string `json:"content"`
	Version      int    `json:"version"`
}

func v2UpdatePage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]

	var req v2PageUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		v2Error(w, "invalid request body", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	existing, err := dao.SelectPage(ctx, id, -1)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	var page datastore.Page
	if existing != nil {
		page = *existing
	}

	page.Name = req.Name
	page.Parent = req.ParentID
	page.Seq = req.Seq
	page.Description = req.Description
	page.SiteTemplate = req.SiteTemplate
	page.PageTemplate = req.PageTemplate
	page.Paging = req.Paging
	page.Deleted = !req.Published
	page.SetTargetVersion(strconv.Itoa(req.Version))

	ps := &datastore.PageSet{
		ID:       id,
		Page:     &page,
		PageData: &datastore.PageData{Content: []byte(req.Content)},
		FileSet:  &datastore.FileSet{},
	}

	if err := dao.PutPage(ctx, ps); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	updated, err := dao.SelectPage(ctx, id, -1)
	if err != nil || updated == nil {
		v2JSON(w, map[string]string{"status": "saved"})
		return
	}
	v2JSON(w, toV2PageRes(updated))
}

func v2DeletePage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.RemovePage(ctx, id); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]string{"status": "deleted"})
}

func v2PublishPage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	if err := logic.PutHTMLs(r.Context(), logic.NewPageInfo(id)); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]string{"status": "published"})
}

func v2DeletePages(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IDs) == 0 {
		v2Error(w, "invalid request body", 400)
		return
	}
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.RemovePages(ctx, req.IDs); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]interface{}{"status": "deleted", "count": len(req.IDs)})
}

func v2PublishPages(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IDs) == 0 {
		v2Error(w, "invalid request body", 400)
		return
	}
	if err := logic.PutHTMLs(r.Context(), logic.NewPageInfos(req.IDs...)...); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]interface{}{"status": "published", "count": len(req.IDs)})
}

func v2UnpublishPage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.RemoveHTML(ctx, id); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]string{"status": "unpublished"})
}

type v2SequenceItem struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Version int    `json:"version"`
}

func v2PageSequence(w http.ResponseWriter, r *http.Request) {
	var items []v2SequenceItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		v2Error(w, "invalid request body", 400)
		return
	}
	if len(items) == 0 {
		v2JSON(w, map[string]string{"status": "ok"})
		return
	}

	ids := make([]string, len(items))
	enables := make([]string, len(items))
	versions := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.ID
		if item.Enabled {
			enables[i] = "true"
		} else {
			enables[i] = "false"
		}
		versions[i] = strconv.Itoa(item.Version)
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.PutPageSequence(ctx,
		strings.Join(ids, ","),
		strings.Join(enables, ","),
		strings.Join(versions, ","),
	); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]string{"status": "ok"})
}

func v2PageSort(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.SortPage(ctx, id); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]string{"status": "ok"})
}

type v2MoveReq struct {
	TargetID string `json:"targetId"`
}

func v2PageMove(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	var req v2MoveReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		v2Error(w, "invalid request body", 400)
		return
	}
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.MovePage(ctx, id, req.TargetID); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]string{"status": "ok"})
}

type v2ImageRes struct {
	URL   string `json:"url"`
	Draft bool   `json:"draft"`
}

func v2GetPageImage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	draftID := datastore.CreateDraftPageImageID(id)
	if dao.ExistFile(ctx, draftID) {
		v2JSON(w, v2ImageRes{URL: "/manage/file/view/" + draftID, Draft: true})
		return
	}
	if dao.ExistFile(ctx, id) {
		v2JSON(w, v2ImageRes{URL: "/manage/file/view/" + id, Draft: false})
		return
	}
	v2JSON(w, v2ImageRes{})
}

func v2UploadPageImage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]

	if err := r.ParseMultipartForm(maxImageSize); err != nil {
		v2Error(w, "リクエストの解析に失敗しました", 400)
		return
	}

	upload, header, err := r.FormFile("file")
	if err != nil {
		v2Error(w, "ファイルが見つかりません", 400)
		return
	}
	defer upload.Close()

	b, converted, err := logic.ConvertImage(upload)
	if err != nil {
		v2Error(w, err.Error(), 400)
		return
	}

	mime := header.Header.Get("Content-Type")
	if converted {
		mime = "image/jpeg"
	}

	var f datastore.File
	var fd datastore.FileData
	f.Size = int64(len(b))
	f.Type = datastore.FileTypePageImage
	fd.Content = b
	fd.Mime = mime

	draftID := datastore.CreateDraftPageImageID(id)
	fs := &datastore.FileSet{
		ID:       draftID,
		Name:     header.Filename,
		File:     &f,
		FileData: &fd,
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.SaveFile(ctx, fs); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	if err := dao.TouchPage(ctx, id); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	v2JSON(w, v2ImageRes{URL: "/manage/file/view/" + draftID, Draft: true})
}

func v2DeletePageImage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	draftID := datastore.CreateDraftPageImageID(id)
	if dao.ExistFile(ctx, draftID) {
		// 下書きを削除して公開画像に戻す
		if err := dao.RemoveFile(ctx, draftID); err != nil {
			v2Error(w, err.Error(), 500)
			return
		}
	} else if dao.ExistFile(ctx, id) {
		// 下書きなし・公開画像のみ → 公開画像を削除
		if err := dao.RemoveFile(ctx, id); err != nil {
			v2Error(w, err.Error(), 500)
			return
		}
	}

	if err := dao.TouchPage(ctx, id); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	// 削除後の現在の画像状態を返す
	draftID2 := datastore.CreateDraftPageImageID(id)
	if dao.ExistFile(ctx, draftID2) {
		v2JSON(w, v2ImageRes{URL: "/manage/file/view/" + draftID2, Draft: true})
		return
	}
	if dao.ExistFile(ctx, id) {
		v2JSON(w, v2ImageRes{URL: "/manage/file/view/" + id, Draft: false})
		return
	}
	v2JSON(w, v2ImageRes{})
}
