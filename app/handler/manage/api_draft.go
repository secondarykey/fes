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

func registerDraftAPI(r *mux.Router) {
	r.HandleFunc("/draft/current", apiGetCurrentDraft).Methods("GET")
	r.HandleFunc("/draft/current/{key}", apiSetCurrentDraft).Methods("POST")
	r.HandleFunc("/draft/", apiListDrafts).Methods("GET")
	r.HandleFunc("/draft/", apiCreateDraft).Methods("POST")
	r.HandleFunc("/draft/{key}", apiGetDraft).Methods("GET")
	r.HandleFunc("/draft/{key}", apiUpdateDraft).Methods("POST")
	r.HandleFunc("/draft/{key}", apiDeleteDraft).Methods("DELETE")
	r.HandleFunc("/draft/{key}/lock", apiToggleDraftLock).Methods("POST")
	r.HandleFunc("/draft/{key}/publish", apiPublishDraft).Methods("POST")
	r.HandleFunc("/draft/{key}/page", apiAddDraftPage).Methods("POST")
	r.HandleFunc("/draft/{key}/pages", apiAddDraftPages).Methods("POST")
	r.HandleFunc("/draft/{key}/page/{pageKey}", apiRemoveDraftPage).Methods("DELETE")
}

// ── レスポンス型 ──────────────────────────────────────────

type apiDraftRes struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Note      string    `json:"note"`
	Lock      bool      `json:"lock"`
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type apiDraftPageRes struct {
	ID            string    `json:"id"`
	PageID        string    `json:"pageId"`
	Name          string    `json:"name"`
	Seq           int       `json:"seq"`
	PublishUpdate bool      `json:"publishUpdate"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func toAPIDraftRes(d *datastore.Draft) apiDraftRes {
	id := ""
	if d.Key != nil {
		id = d.Key.Name
	}
	return apiDraftRes{ID: id, Name: d.Name, Note: d.Note, Lock: d.Lock, Version: d.Version, UpdatedAt: d.UpdatedAt}
}

func toAPIDraftPageRes(p *datastore.DraftPage) apiDraftPageRes {
	id := ""
	if p.Key != nil {
		id = p.Key.Name
	}
	return apiDraftPageRes{
		ID:            id,
		PageID:        p.PageID,
		Name:          p.Name,
		Seq:           p.Seq,
		PublishUpdate: p.PublishUpdate,
		UpdatedAt:     p.UpdatedAt,
	}
}

func draftSetToRes(set *datastore.DraftSet) map[string]interface{} {
	pages := make([]apiDraftPageRes, len(set.Pages))
	for i, p := range set.Pages {
		pages[i] = toAPIDraftPageRes(p)
	}
	return map[string]interface{}{
		"draft": toAPIDraftRes(set.Draft),
		"pages": pages,
	}
}

// ── ハンドラ ──────────────────────────────────────────────

func apiGetCurrentDraft(w http.ResponseWriter, r *http.Request) {
	id, _ := GetDraftId(r)
	if id == "" {
		apiJSON(w, map[string]interface{}{"id": nil})
		return
	}
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	draft, err := dao.SelectDraft(ctx, id)
	if err != nil || draft == nil {
		// セッションに残っているが実体がない場合はクリア
		SetDraftId(w, r, "")
		apiJSON(w, map[string]interface{}{"id": nil})
		return
	}
	apiJSON(w, toAPIDraftRes(draft))
}

func apiSetCurrentDraft(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	if err := SetDraftId(w, r, id); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"status": "ok", "id": id})
}

func apiListDrafts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	cursor := r.URL.Query().Get("cursor")
	drafts, nextCursor, err := dao.SelectDrafts(ctx, cursor)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	res := make([]apiDraftRes, len(drafts))
	for i := range drafts {
		res[i] = toAPIDraftRes(&drafts[i])
	}
	apiJSON(w, map[string]interface{}{"drafts": res, "nextCursor": nextCursor})
}

type apiDraftCreateReq struct {
	Name string `json:"name"`
}

func apiCreateDraft(w http.ResponseWriter, r *http.Request) {
	var req apiDraftCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, "invalid request body", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	key := datastore.CreateDraftKey()
	draft := &datastore.Draft{Name: req.Name, Lock: true}
	draft.LoadKey(key)

	// 新規なので existSet.Pages は空、forms も空 → copyDraft([], []) で OK
	set := &datastore.DraftSet{Draft: draft, Pages: []*datastore.DraftPage{}}
	if err := dao.PutDraftSet(ctx, set); err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	apiJSON(w, toAPIDraftRes(draft))
}

func apiGetDraft(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	set, err := dao.SelectDraftSet(ctx, id)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	if set.Draft == nil {
		apiError(w, "draft not found", 404)
		return
	}
	apiJSON(w, draftSetToRes(set))
}

type apiDraftUpdateReq struct {
	Name    string `json:"name"`
	Note    string `json:"note"`
	Lock    bool   `json:"lock"`
	Version int    `json:"version"`
	Pages   []struct {
		ID            string `json:"id"`
		PublishUpdate bool   `json:"publishUpdate"`
	} `json:"pages"`
}

func apiUpdateDraft(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]

	var req apiDraftUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, "invalid request body", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	set, err := dao.SelectDraftSet(ctx, id)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	if set.Draft == nil {
		apiError(w, "draft not found", 404)
		return
	}

	set.Draft.Name = req.Name
	set.Draft.Note = req.Note
	set.Draft.Lock = req.Lock
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
		apiError(w, err.Error(), 500)
		return
	}

	updated, err := dao.SelectDraftSet(ctx, id)
	if err != nil || updated.Draft == nil {
		apiJSON(w, map[string]string{"status": "saved"})
		return
	}
	apiJSON(w, draftSetToRes(updated))
}

func apiDeleteDraft(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.RemoveDraft(ctx, id); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"status": "deleted"})
}

func apiToggleDraftLock(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	draft, err := dao.SelectDraft(ctx, id)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	if draft == nil {
		apiError(w, "draft not found", 404)
		return
	}

	draft.Lock = !draft.Lock
	if err := dao.PutDraft(ctx, draft); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, toAPIDraftRes(draft))
}

func apiPublishDraft(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	draft, err := dao.SelectDraft(ctx, id)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	if draft == nil {
		apiError(w, "draft not found", 404)
		return
	}
	if draft.Lock {
		apiError(w, "現在作業中の為公開できません", 400)
		return
	}

	pages, err := dao.SelectDraftPages(ctx, id)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	if len(pages) == 0 {
		apiError(w, "公開するページがありません", 400)
		return
	}

	infos := make([]*logic.PageInfo, len(pages))
	for i, p := range pages {
		info := logic.NewPageInfo(p.PageID)
		info.Publish = p.PublishUpdate
		infos[i] = info
	}

	if err := logic.PutHTMLs(ctx, infos...); err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	if err := dao.RemoveDraft(ctx, id); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"status": "published"})
}

type apiAddDraftPageReq struct {
	PageID string `json:"pageId"`
}

func apiAddDraftPage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]

	var req apiAddDraftPageReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, "invalid request body", 400)
		return
	}
	if req.PageID == "" {
		apiError(w, "pageId は必須です", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.AddDraftPage(ctx, id, req.PageID); err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	pages, err := dao.SelectDraftPages(ctx, id)
	if err != nil {
		apiJSON(w, map[string]string{"status": "added"})
		return
	}
	res := make([]apiDraftPageRes, len(pages))
	for i, p := range pages {
		res[i] = toAPIDraftPageRes(p)
	}
	apiJSON(w, res)
}

type apiAddDraftPagesReq struct {
	PageIDs []string `json:"pageIds"`
}

func apiAddDraftPages(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]

	var req apiAddDraftPagesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.PageIDs) == 0 {
		apiError(w, "invalid request body", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	for _, pageID := range req.PageIDs {
		if err := dao.AddDraftPage(ctx, id, pageID); err != nil {
			apiError(w, err.Error(), 500)
			return
		}
	}

	pages, err := dao.SelectDraftPages(ctx, id)
	if err != nil {
		apiJSON(w, map[string]string{"status": "added"})
		return
	}
	res := make([]apiDraftPageRes, len(pages))
	for i, p := range pages {
		res[i] = toAPIDraftPageRes(p)
	}
	apiJSON(w, res)
}

func apiRemoveDraftPage(w http.ResponseWriter, r *http.Request) {
	pageKey := mux.Vars(r)["pageKey"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if _, err := dao.RemoveDraftPage(ctx, pageKey); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"status": "removed"})
}
