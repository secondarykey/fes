package manage

import (
	"app/datastore"
	"app/logic"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func registerV2DraftAPI(r *mux.Router) {
	r.HandleFunc("/draft/current", v2GetCurrentDraft).Methods("GET")
	r.HandleFunc("/draft/current/{key}", v2SetCurrentDraft).Methods("POST")
	r.HandleFunc("/draft/", v2ListDrafts).Methods("GET")
	r.HandleFunc("/draft/", v2CreateDraft).Methods("POST")
	r.HandleFunc("/draft/{key}", v2GetDraft).Methods("GET")
	r.HandleFunc("/draft/{key}", v2UpdateDraft).Methods("POST")
	r.HandleFunc("/draft/{key}", v2DeleteDraft).Methods("DELETE")
	r.HandleFunc("/draft/{key}/publish", v2PublishDraft).Methods("POST")
	r.HandleFunc("/draft/{key}/page", v2AddDraftPage).Methods("POST")
	r.HandleFunc("/draft/{key}/pages", v2AddDraftPages).Methods("POST")
	r.HandleFunc("/draft/{key}/page/{pageKey}", v2RemoveDraftPage).Methods("DELETE")
}

// ── レスポンス型 ──────────────────────────────────────────

type v2DraftRes struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type v2DraftPageRes struct {
	ID            string `json:"id"`
	PageID        string `json:"pageId"`
	Name          string `json:"name"`
	Seq           int    `json:"seq"`
	PublishUpdate bool   `json:"publishUpdate"`
}

func toV2DraftRes(d *datastore.Draft) v2DraftRes {
	id := ""
	if d.Key != nil {
		id = d.Key.Name
	}
	return v2DraftRes{ID: id, Name: d.Name, Version: d.Version, UpdatedAt: d.UpdatedAt}
}

func toV2DraftPageRes(p *datastore.DraftPage) v2DraftPageRes {
	id := ""
	if p.Key != nil {
		id = p.Key.Name
	}
	return v2DraftPageRes{
		ID:            id,
		PageID:        p.PageID,
		Name:          p.Name,
		Seq:           p.Seq,
		PublishUpdate: p.PublishUpdate,
	}
}

func draftSetToRes(set *datastore.DraftSet) map[string]interface{} {
	pages := make([]v2DraftPageRes, len(set.Pages))
	for i, p := range set.Pages {
		pages[i] = toV2DraftPageRes(p)
	}
	return map[string]interface{}{
		"draft": toV2DraftRes(set.Draft),
		"pages": pages,
	}
}

// ── ハンドラ ──────────────────────────────────────────────

func v2GetCurrentDraft(w http.ResponseWriter, r *http.Request) {
	id, _ := GetDraftId(r)
	if id == "" {
		v2JSON(w, map[string]interface{}{"id": nil})
		return
	}
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	draft, err := dao.SelectDraft(ctx, id)
	if err != nil || draft == nil {
		// セッションに残っているが実体がない場合はクリア
		SetDraftId(w, r, "")
		v2JSON(w, map[string]interface{}{"id": nil})
		return
	}
	v2JSON(w, toV2DraftRes(draft))
}

func v2SetCurrentDraft(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	if err := SetDraftId(w, r, id); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]string{"status": "ok", "id": id})
}

func v2ListDrafts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	cursor := r.URL.Query().Get("cursor")
	drafts, nextCursor, err := dao.SelectDrafts(ctx, cursor)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	res := make([]v2DraftRes, len(drafts))
	for i := range drafts {
		res[i] = toV2DraftRes(&drafts[i])
	}
	v2JSON(w, map[string]interface{}{"drafts": res, "nextCursor": nextCursor})
}

type v2DraftCreateReq struct {
	Name string `json:"name"`
}

func v2CreateDraft(w http.ResponseWriter, r *http.Request) {
	var req v2DraftCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		v2Error(w, "invalid request body", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	key := datastore.CreateDraftKey()
	draft := &datastore.Draft{Name: req.Name}
	draft.LoadKey(key)

	// 新規なので existSet.Pages は空、forms も空 → copyDraft([], []) で OK
	set := &datastore.DraftSet{Draft: draft, Pages: []*datastore.DraftPage{}}
	if err := dao.PutDraftSet(ctx, set); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	v2JSON(w, toV2DraftRes(draft))
}

func v2GetDraft(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	set, err := dao.SelectDraftSet(ctx, id)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	if set.Draft == nil {
		v2Error(w, "draft not found", 404)
		return
	}
	v2JSON(w, draftSetToRes(set))
}

type v2DraftUpdateReq struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
	Pages   []struct {
		ID            string `json:"id"`
		PublishUpdate bool   `json:"publishUpdate"`
	} `json:"pages"`
}

func v2UpdateDraft(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]

	var req v2DraftUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		v2Error(w, "invalid request body", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	set, err := dao.SelectDraftSet(ctx, id)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	if set.Draft == nil {
		v2Error(w, "draft not found", 404)
		return
	}

	set.Draft.Name = req.Name
	set.Draft.SetTargetVersion(fmt.Sprintf("%d", req.Version))

	// req.Pages が既存ページと件数一致する場合のみ順序・PublishUpdate を更新
	if len(req.Pages) == len(set.Pages) {
		forms := make([]*datastore.DraftPage, len(req.Pages))
		for i, rp := range req.Pages {
			dp := &datastore.DraftPage{}
			dp.LoadKey(datastore.GetDraftPageKey(rp.ID))
			dp.Seq = i + 1
			dp.PublishUpdate = rp.PublishUpdate
			forms[i] = dp
		}
		set.Pages = forms
	}
	// 件数不一致の場合は set.Pages をそのまま渡す（名前だけ更新）

	if err := dao.PutDraftSet(ctx, set); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	updated, err := dao.SelectDraftSet(ctx, id)
	if err != nil || updated.Draft == nil {
		v2JSON(w, map[string]string{"status": "saved"})
		return
	}
	v2JSON(w, draftSetToRes(updated))
}

func v2DeleteDraft(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.RemoveDraft(ctx, id); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]string{"status": "deleted"})
}

func v2PublishDraft(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	pages, err := dao.SelectDraftPages(ctx, id)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	if len(pages) == 0 {
		v2Error(w, "公開するページがありません", 400)
		return
	}

	infos := make([]*logic.PageInfo, len(pages))
	for i, p := range pages {
		info := logic.NewPageInfo(p.PageID)
		info.Publish = p.PublishUpdate
		infos[i] = info
	}

	if err := logic.PutHTMLs(ctx, infos...); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	if err := dao.RemoveDraft(ctx, id); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]string{"status": "published"})
}

type v2AddDraftPageReq struct {
	PageID string `json:"pageId"`
}

func v2AddDraftPage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]

	var req v2AddDraftPageReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		v2Error(w, "invalid request body", 400)
		return
	}
	if req.PageID == "" {
		v2Error(w, "pageId は必須です", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.AddDraftPage(ctx, id, req.PageID); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	pages, err := dao.SelectDraftPages(ctx, id)
	if err != nil {
		v2JSON(w, map[string]string{"status": "added"})
		return
	}
	res := make([]v2DraftPageRes, len(pages))
	for i, p := range pages {
		res[i] = toV2DraftPageRes(p)
	}
	v2JSON(w, res)
}

type v2AddDraftPagesReq struct {
	PageIDs []string `json:"pageIds"`
}

func v2AddDraftPages(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]

	var req v2AddDraftPagesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.PageIDs) == 0 {
		v2Error(w, "invalid request body", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	for _, pageID := range req.PageIDs {
		if err := dao.AddDraftPage(ctx, id, pageID); err != nil {
			v2Error(w, err.Error(), 500)
			return
		}
	}

	pages, err := dao.SelectDraftPages(ctx, id)
	if err != nil {
		v2JSON(w, map[string]string{"status": "added"})
		return
	}
	res := make([]v2DraftPageRes, len(pages))
	for i, p := range pages {
		res[i] = toV2DraftPageRes(p)
	}
	v2JSON(w, res)
}

func v2RemoveDraftPage(w http.ResponseWriter, r *http.Request) {
	pageKey := mux.Vars(r)["pageKey"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if _, err := dao.RemoveDraftPage(ctx, pageKey); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]string{"status": "removed"})
}
