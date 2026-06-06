package manage

import (
	"app/datastore"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func registerVariableAPI(r *mux.Router) {
	r.HandleFunc("/variable/", apiListVariables).Methods("GET")
	r.HandleFunc("/variable/", apiCreateVariable).Methods("POST")
	r.HandleFunc("/variable/{key}", apiGetVariable).Methods("GET")
	r.HandleFunc("/variable/{key}", apiUpdateVariable).Methods("POST")
	r.HandleFunc("/variable/{key}", apiDeleteVariable).Methods("DELETE")
}

type apiVariableRes struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toAPIVariableRes(v *datastore.Variable, content string) apiVariableRes {
	id := ""
	if v.Key != nil {
		id = v.Key.Name
	}
	return apiVariableRes{
		ID:        id,
		Content:   content,
		Version:   v.Version,
		UpdatedAt: v.UpdatedAt,
	}
}

func apiListVariables(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	cursor := r.URL.Query().Get("cursor")
	variables, nextCursor, err := dao.SelectVariables(ctx, cursor)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	res := make([]apiVariableRes, len(variables))
	for i, v := range variables {
		id := ""
		if v.Key != nil {
			id = v.Key.Name
		}
		res[i] = apiVariableRes{ID: id, Version: v.Version, UpdatedAt: v.UpdatedAt}
	}

	apiJSON(w, map[string]interface{}{
		"variables":  res,
		"nextCursor": nextCursor,
	})
}

func apiGetVariable(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	v, err := dao.SelectVariable(ctx, id)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	if v == nil {
		apiError(w, "variable not found", 404)
		return
	}

	vd, err := dao.SelectVariableData(ctx, id)
	content := ""
	if err == nil && vd != nil {
		content = string(vd.Content)
	}

	apiJSON(w, toAPIVariableRes(v, content))
}

type apiVariableCreateReq struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

func apiCreateVariable(w http.ResponseWriter, r *http.Request) {
	var req apiVariableCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, "invalid request body", 400)
		return
	}
	if req.ID == "" {
		apiError(w, "id は必須です", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	existing, err := dao.SelectVariable(ctx, req.ID)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	if existing != nil {
		apiError(w, fmt.Sprintf("変数 '%s' は既に存在します", req.ID), 409)
		return
	}

	vs := &datastore.VariableSet{
		ID:           req.ID,
		Variable:     &datastore.Variable{},
		VariableData: &datastore.VariableData{Content: []byte(req.Content)},
	}
	if err := dao.PutVariable(ctx, vs); err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	saved, err := dao.SelectVariable(ctx, req.ID)
	if err != nil || saved == nil {
		apiJSON(w, map[string]string{"status": "created"})
		return
	}
	apiJSON(w, toAPIVariableRes(saved, req.Content))
}

type apiVariableUpdateReq struct {
	Content string `json:"content"`
	Version int    `json:"version"`
}

func apiUpdateVariable(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]

	var req apiVariableUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, "invalid request body", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	existing, err := dao.SelectVariable(ctx, id)
	if err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	var v datastore.Variable
	if existing != nil {
		v = *existing
	}
	v.SetTargetVersion(fmt.Sprintf("%d", req.Version))

	vs := &datastore.VariableSet{
		ID:           id,
		Variable:     &v,
		VariableData: &datastore.VariableData{Content: []byte(req.Content)},
	}
	if err := dao.PutVariable(ctx, vs); err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	updated, err := dao.SelectVariable(ctx, id)
	if err != nil || updated == nil {
		apiJSON(w, map[string]string{"status": "saved"})
		return
	}
	apiJSON(w, toAPIVariableRes(updated, req.Content))
}

func apiDeleteVariable(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.RemoveVariable(ctx, id); err != nil {
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, map[string]string{"status": "deleted"})
}
