package manage

import (
	"app/datastore"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func registerV2TemplateAPI(r *mux.Router) {
	r.HandleFunc("/template/", v2ListTemplates).Methods("GET")
	r.HandleFunc("/template/sequence", v2TemplateSequence).Methods("POST")
	r.HandleFunc("/template/new", v2NewTemplate).Methods("GET")
	r.HandleFunc("/template/{key}/references", v2TemplateReferences).Methods("GET")
	r.HandleFunc("/template/{key}", v2GetTemplate).Methods("GET")
	r.HandleFunc("/template/{key}", v2UpdateTemplate).Methods("POST")
	r.HandleFunc("/template/{key}", v2DeleteTemplate).Methods("DELETE")
}

type v2TemplateDetailRes struct {
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

func toV2TemplateDetailRes(t *datastore.Template, content string) v2TemplateDetailRes {
	id := ""
	if t.Key != nil {
		id = t.Key.Name
	}
	return v2TemplateDetailRes{
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

func v2ListTemplates(w http.ResponseWriter, r *http.Request) {
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
		v2Error(w, err.Error(), 500)
		return
	}

	res := make([]v2TemplateRes, len(templates))
	for i, t := range templates {
		id := ""
		if t.Key != nil {
			id = t.Key.Name
		}
		res[i] = v2TemplateRes{ID: id, Name: t.Name, Type: t.Type}
	}

	v2JSON(w, map[string]interface{}{
		"templates":  res,
		"nextCursor": nextCursor,
	})
}

func v2NewTemplate(w http.ResponseWriter, r *http.Request) {
	key := datastore.CreateTemplateKey()
	t := &datastore.Template{}
	t.LoadKey(key)
	v2JSON(w, toV2TemplateDetailRes(t, ""))
}

func v2GetTemplate(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	t, err := dao.SelectTemplate(ctx, id)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	if t == nil {
		v2Error(w, "template not found", 404)
		return
	}

	td, err := dao.SelectTemplateData(ctx, id)
	content := ""
	if err == nil && td != nil {
		content = string(td.Content)
	}

	v2JSON(w, toV2TemplateDetailRes(t, content))
}

type v2TemplateUpdateReq struct {
	Name    string `json:"name"`
	Type    int    `json:"type"`
	Content string `json:"content"`
	Version int    `json:"version"`
}

func v2UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]

	var req v2TemplateUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		v2Error(w, "invalid request body", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	existing, err := dao.SelectTemplate(ctx, id)
	if err != nil {
		v2Error(w, err.Error(), 500)
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
		v2Error(w, err.Error(), 500)
		return
	}

	updated, err := dao.SelectTemplate(ctx, id)
	if err != nil || updated == nil {
		v2JSON(w, map[string]string{"status": "saved"})
		return
	}
	v2JSON(w, toV2TemplateDetailRes(updated, req.Content))
}

func v2DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	ok, err := dao.UsingTemplate(ctx, id)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	if ok {
		v2Error(w, "このテンプレートはページで使用中のため削除できません", 409)
		return
	}

	if err := dao.RemoveTemplate(ctx, id); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]string{"status": "deleted"})
}

func v2TemplateReferences(w http.ResponseWriter, r *http.Request) {
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
			v2Error(w, err.Error(), 500)
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
	v2JSON(w, map[string]interface{}{"pages": pages})
}

type v2TemplateSeqReqItem struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}

func v2TemplateSequence(w http.ResponseWriter, r *http.Request) {
	var req []v2TemplateSeqReqItem
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		v2Error(w, "invalid request body", 400)
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
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]string{"status": "ok"})
}
