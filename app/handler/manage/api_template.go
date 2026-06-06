package manage

import (
	"app/datastore"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func registerTemplateAPI(r *mux.Router) {
	r.HandleFunc("/template/", apiListTemplates).Methods("GET")
	r.HandleFunc("/template/sequence", apiTemplateSequence).Methods("POST")
	r.HandleFunc("/template/new", apiNewTemplate).Methods("GET")
	r.HandleFunc("/template/{key}/references", apiTemplateReferences).Methods("GET")
	r.HandleFunc("/template/{key}", apiGetTemplate).Methods("GET")
	r.HandleFunc("/template/{key}", apiUpdateTemplate).Methods("POST")
	r.HandleFunc("/template/{key}", apiDeleteTemplate).Methods("DELETE")
}

type apiTemplateDetailRes struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      int       `json:"type"`
	Seq       int       `json:"seq"`
	Content   string    `json:"content"`
	Version   int       `json:"version"`
	Deleted   bool      `json:"deleted"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toAPITemplateDetailRes(t *datastore.Template, content string) apiTemplateDetailRes {
	id := ""
	if t.Key != nil {
		id = t.Key.Name
	}
	return apiTemplateDetailRes{
		ID:        id,
		Name:      t.Name,
		Type:      t.Type,
		Seq:       t.Seq,
		Content:   content,
		Version:   t.Version,
		Deleted:   t.Deleted,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func apiListTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	typ := r.URL.Query().Get("type")
	if typ == "" {
		typ = "all"
	}
	cursor := r.URL.Query().Get("cursor")

	templates, nextCursor, err := dao.SelectTemplates(ctx, typ, cursor)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	res := make([]apiTemplateRes, len(templates))
	for i, t := range templates {
		id := ""
		if t.Key != nil {
			id = t.Key.Name
		}
		res[i] = apiTemplateRes{ID: id, Name: t.Name, Type: t.Type}
	}

	apiJSON(w, map[string]interface{}{
		"templates":  res,
		"nextCursor": nextCursor,
	})
}

func apiNewTemplate(w http.ResponseWriter, r *http.Request) {
	key := datastore.CreateTemplateKey()
	t := &datastore.Template{}
	t.LoadKey(key)
	apiJSON(w, toAPITemplateDetailRes(t, ""))
}

func apiGetTemplate(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	t, err := dao.SelectTemplate(ctx, id)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	if t == nil {
		apiError(w, "template not found", 404)
		return
	}

	td, err := dao.SelectTemplateData(ctx, id)
	content := ""
	if err == nil && td != nil {
		content = string(td.Content)
	}

	apiJSON(w, toAPITemplateDetailRes(t, content))
}

type apiTemplateUpdateReq struct {
	Name    string `json:"name"`
	Type    int    `json:"type"`
	Content string `json:"content"`
	Version int    `json:"version"`
}

func apiUpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]

	var req apiTemplateUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, "invalid request body", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	existing, err := dao.SelectTemplate(ctx, id)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	var t datastore.Template
	if existing != nil {
		t = *existing
	}
	t.LoadKey(datastore.GetTemplateKey(id))

	t.Name = req.Name
	t.Type = req.Type
	t.SetTargetVersion(fmt.Sprintf("%d", req.Version))

	td := &datastore.TemplateData{Content: []byte(req.Content)}
	td.LoadKey(datastore.GetTemplateDataKey(id))

	ts := &datastore.TemplateSet{
		ID:           id,
		Template:     &t,
		TemplateData: td,
	}

	if err := dao.PutTemplate(ctx, ts); err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	updated, err := dao.SelectTemplate(ctx, id)
	if err != nil || updated == nil {
		apiJSON(w, map[string]string{"status": "saved"})
		return
	}
	apiJSON(w, toAPITemplateDetailRes(updated, req.Content))
}

func apiDeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	ok, err := dao.UsingTemplate(ctx, id)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	if ok {
		apiError(w, "このテンプレートはページで使用中のため削除できません", 409)
		return
	}

	if err := dao.RemoveTemplate(ctx, id); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"status": "deleted"})
}

func apiTemplateReferences(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	// PageTemplate(2) と SiteTemplate(1) の両方を取得してマージ
	type refPage struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		UpdatedAt time.Time `json:"updatedAt"`
	}

	seen := map[string]bool{}
	var pages []refPage

	for _, typ := range []int{2, 1} {
		ps, err := dao.SelectReferencePages(ctx, id, typ)
		if err != nil {
			apiError(w, err.Error(), 500)
			return
		}
		for _, p := range ps {
			pid := p.Key.Name
			if seen[pid] || p.Deleted {
				continue
			}
			seen[pid] = true
			pages = append(pages, refPage{
				ID:        pid,
				Name:      p.Name,
				UpdatedAt: p.UpdatedAt,
			})
		}
	}

	if pages == nil {
		pages = []refPage{}
	}
	apiJSON(w, map[string]interface{}{"pages": pages})
}

type apiTemplateSeqReqItem struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}

func apiTemplateSequence(w http.ResponseWriter, r *http.Request) {
	var req []apiTemplateSeqReqItem
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, "invalid request body", 400)
		return
	}

	items := make([]datastore.TemplateSeqItem, len(req))
	for i, r := range req {
		items[i] = datastore.TemplateSeqItem{
			ID:      r.ID,
			Version: fmt.Sprintf("%d", r.Version),
		}
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.PutTemplateSequence(ctx, items); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"status": "ok"})
}
