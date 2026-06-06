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

func registerPageAPI(r *mux.Router) {
	r.HandleFunc("/page/", apiGetRootPage).Methods("GET")
	// /page/new/{parentKey} は /page/{key} より先に登録する
	r.HandleFunc("/page/new/{parentKey}", apiNewPage).Methods("GET")
	r.HandleFunc("/page/children/{key}", apiGetChildren).Methods("GET")
	r.HandleFunc("/page/{key}/image", apiGetPageImage).Methods("GET")
	r.HandleFunc("/page/{key}/image", apiUploadPageImage).Methods("POST")
	r.HandleFunc("/page/{key}/image", apiDeletePageImage).Methods("DELETE")
	r.HandleFunc("/page/{key}/sequence", apiPageSequence).Methods("POST")
	r.HandleFunc("/page/{key}/sort", apiPageSort).Methods("POST")
	r.HandleFunc("/page/{key}/move", apiPageMove).Methods("POST")
	r.HandleFunc("/page/{key}", apiGetPage).Methods("GET")
	r.HandleFunc("/page/{key}", apiUpdatePage).Methods("POST")
	r.HandleFunc("/page/{key}", apiDeletePage).Methods("DELETE")
	r.HandleFunc("/page/", apiDeletePages).Methods("DELETE")
	r.HandleFunc("/html/publish/{key}", apiPublishPage).Methods("POST")
	r.HandleFunc("/html/unpublish/{key}", apiUnpublishPage).Methods("POST")
	r.HandleFunc("/html/publish-pages", apiPublishPages).Methods("POST")
}

// ── レスポンス型 ──────────────────────────────────────────

type apiPageRes struct {
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

type apiTemplateRes struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"`
	Seq  int    `json:"seq"`
}

func toAPIPageRes(p *datastore.Page) apiPageRes {
	if p == nil {
		return apiPageRes{}
	}
	id := ""
	if p.Key != nil {
		id = p.Key.Name
	}
	canPublish := !p.Deleted && p.UpdatedAt.Unix() > p.Republish.Unix()+1
	return apiPageRes{
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

func apiJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func apiError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func apiTemplates(r *http.Request, dao *datastore.Dao) ([]apiTemplateRes, error) {
	templates, _, err := dao.SelectTemplates(r.Context(), "all", datastore.NoLimitCursor)
	if err != nil {
		return nil, err
	}
	res := make([]apiTemplateRes, len(templates))
	for i, t := range templates {
		res[i] = apiTemplateRes{ID: t.Key.Name, Name: t.Name, Type: t.Type, Seq: t.Seq}
	}
	return res, nil
}

// ── ハンドラ ──────────────────────────────────────────────

func apiGetRootPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	page, err := dao.SelectRootPage(ctx)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	if page == nil {
		apiError(w, "root page not found", 404)
		return
	}

	children, _, err := dao.SelectChildrenPage(ctx, page.Key.Name, datastore.NoLimitCursor, 0, true)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	childRes := make([]apiPageRes, len(children))
	for i := range children {
		childRes[i] = toAPIPageRes(&children[i])
	}

	apiJSON(w, map[string]interface{}{"page": toAPIPageRes(page), "children": childRes})
}

func apiGetChildren(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	children, _, err := dao.SelectChildrenPage(ctx, id, datastore.NoLimitCursor, 0, true)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	res := make([]apiPageRes, len(children))
	for i := range children {
		res[i] = toAPIPageRes(&children[i])
	}
	apiJSON(w, res)
}

func apiNewPage(w http.ResponseWriter, r *http.Request) {
	parentKey := mux.Vars(r)["parentKey"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	page := &datastore.Page{}
	page.Parent = parentKey
	page.Deleted = false
	page.LoadKey(datastore.CreatePageKey())

	// パンくずリスト（親を辿る）
	var breadcrumbs []apiPageRes
	cur := parentKey
	for cur != "" {
		pp, err := dao.SelectPage(ctx, cur, -1)
		if err != nil || pp == nil {
			break
		}
		breadcrumbs = append([]apiPageRes{toAPIPageRes(pp)}, breadcrumbs...)
		cur = pp.Parent
	}

	tmplRes, err := apiTemplates(r, dao)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	apiJSON(w, map[string]interface{}{
		"page":        toAPIPageRes(page),
		"pageData":    "",
		"children":    []apiPageRes{},
		"breadcrumbs": breadcrumbs,
		"templates":   tmplRes,
	})
}

func apiGetPage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	page, err := dao.SelectPage(ctx, id, -1)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	if page == nil {
		apiError(w, "page not found", 404)
		return
	}

	pageData, err := dao.SelectPageData(ctx, id)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	content := ""
	if pageData != nil {
		content = string(pageData.Content)
	}

	children, _, err := dao.SelectChildrenPage(ctx, id, datastore.NoLimitCursor, 0, true)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	childRes := make([]apiPageRes, len(children))
	for i := range children {
		childRes[i] = toAPIPageRes(&children[i])
	}

	// パンくずリスト
	var breadcrumbs []apiPageRes
	parent := page.Parent
	for parent != "" {
		pp, err := dao.SelectPage(ctx, parent, -1)
		if err != nil || pp == nil {
			break
		}
		breadcrumbs = append([]apiPageRes{toAPIPageRes(pp)}, breadcrumbs...)
		parent = pp.Parent
	}

	tmplRes, err := apiTemplates(r, dao)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	apiJSON(w, map[string]interface{}{
		"page":        toAPIPageRes(page),
		"pageData":    content,
		"children":    childRes,
		"breadcrumbs": breadcrumbs,
		"templates":   tmplRes,
	})
}

type apiPageUpdateReq struct {
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

func apiUpdatePage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]

	var req apiPageUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, "invalid request body", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	existing, err := dao.SelectPage(ctx, id, -1)
	if err != nil {
		apiError(w, err.Error(), 500)
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
		apiError(w, err.Error(), 500)
		return
	}

	updated, err := dao.SelectPage(ctx, id, -1)
	if err != nil || updated == nil {
		apiJSON(w, map[string]string{"status": "saved"})
		return
	}
	apiJSON(w, toAPIPageRes(updated))
}

func apiDeletePage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.RemovePage(ctx, id); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"status": "deleted"})
}

func apiPublishPage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	if err := logic.PutHTMLs(r.Context(), logic.NewPageInfo(id)); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"status": "published"})
}

func apiDeletePages(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IDs) == 0 {
		apiError(w, "invalid request body", 400)
		return
	}
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.RemovePages(ctx, req.IDs); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]interface{}{"status": "deleted", "count": len(req.IDs)})
}

func apiPublishPages(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IDs) == 0 {
		apiError(w, "invalid request body", 400)
		return
	}
	if err := logic.PutHTMLs(r.Context(), logic.NewPageInfos(req.IDs...)...); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]interface{}{"status": "published", "count": len(req.IDs)})
}

func apiUnpublishPage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.RemoveHTML(ctx, id); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"status": "unpublished"})
}

type apiSequenceItem struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Version int    `json:"version"`
}

func apiPageSequence(w http.ResponseWriter, r *http.Request) {
	var items []apiSequenceItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		apiError(w, "invalid request body", 400)
		return
	}
	if len(items) == 0 {
		apiJSON(w, map[string]string{"status": "ok"})
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
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"status": "ok"})
}

func apiPageSort(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.SortPage(ctx, id); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"status": "ok"})
}

type apiMoveReq struct {
	TargetID string `json:"targetId"`
}

func apiPageMove(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	var req apiMoveReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, "invalid request body", 400)
		return
	}
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.MovePage(ctx, id, req.TargetID); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"status": "ok"})
}

type apiImageRes struct {
	URL   string `json:"url"`
	Draft bool   `json:"draft"`
}

func apiGetPageImage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	draftID := datastore.CreateDraftPageImageID(id)
	if dao.ExistFile(ctx, draftID) {
		apiJSON(w, apiImageRes{URL: "/manage/v1/file/view/" + draftID, Draft: true})
		return
	}
	if dao.ExistFile(ctx, id) {
		apiJSON(w, apiImageRes{URL: "/manage/v1/file/view/" + id, Draft: false})
		return
	}
	apiJSON(w, apiImageRes{})
}

func apiUploadPageImage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]

	if err := r.ParseMultipartForm(maxImageSize); err != nil {
		apiError(w, "リクエストの解析に失敗しました", 400)
		return
	}

	upload, header, err := r.FormFile("file")
	if err != nil {
		apiError(w, "ファイルが見つかりません", 400)
		return
	}
	defer upload.Close()

	b, converted, err := logic.ConvertImage(upload)
	if err != nil {
		apiError(w, err.Error(), 400)
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
		apiError(w, err.Error(), 500)
		return
	}

	if err := dao.TouchPage(ctx, id); err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	apiJSON(w, apiImageRes{URL: "/manage/v1/file/view/" + draftID, Draft: true})
}

func apiDeletePageImage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	draftID := datastore.CreateDraftPageImageID(id)
	if dao.ExistFile(ctx, draftID) {
		// 下書きを削除して公開画像に戻す
		if err := dao.RemoveFile(ctx, draftID); err != nil {
			apiError(w, err.Error(), 500)
			return
		}
	} else if dao.ExistFile(ctx, id) {
		// 下書きなし・公開画像のみ → 公開画像を削除
		if err := dao.RemoveFile(ctx, id); err != nil {
			apiError(w, err.Error(), 500)
			return
		}
	}

	if err := dao.TouchPage(ctx, id); err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	// 削除後の現在の画像状態を返す
	draftID2 := datastore.CreateDraftPageImageID(id)
	if dao.ExistFile(ctx, draftID2) {
		apiJSON(w, apiImageRes{URL: "/manage/v1/file/view/" + draftID2, Draft: true})
		return
	}
	if dao.ExistFile(ctx, id) {
		apiJSON(w, apiImageRes{URL: "/manage/v1/file/view/" + id, Draft: false})
		return
	}
	apiJSON(w, apiImageRes{})
}
