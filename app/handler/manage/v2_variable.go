package manage

import (
	"app/datastore"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func registerV2VariableAPI(r *mux.Router) {
	r.HandleFunc("/variable/", v2ListVariables).Methods("GET")
	r.HandleFunc("/variable/", v2CreateVariable).Methods("POST")
	r.HandleFunc("/variable/{key}", v2GetVariable).Methods("GET")
	r.HandleFunc("/variable/{key}", v2UpdateVariable).Methods("POST")
	r.HandleFunc("/variable/{key}", v2DeleteVariable).Methods("DELETE")
}

type v2VariableRes struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toV2VariableRes(v *datastore.Variable, content string) v2VariableRes {
	id := ""
	if v.Key != nil {
		id = v.Key.Name
	}
	return v2VariableRes{
		ID:        id,
		Content:   content,
		Version:   v.Version,
		UpdatedAt: v.UpdatedAt,
	}
}

func v2ListVariables(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	cursor := r.URL.Query().Get("cursor")
	variables, nextCursor, err := dao.SelectVariables(ctx, cursor)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	res := make([]v2VariableRes, len(variables))
	for i, v := range variables {
		id := ""
		if v.Key != nil {
			id = v.Key.Name
		}
		res[i] = v2VariableRes{ID: id, Version: v.Version, UpdatedAt: v.UpdatedAt}
	}

	v2JSON(w, map[string]interface{}{
		"variables":  res,
		"nextCursor": nextCursor,
	})
}

func v2GetVariable(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	v, err := dao.SelectVariable(ctx, id)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	if v == nil {
		v2Error(w, "variable not found", 404)
		return
	}

	vd, err := dao.SelectVariableData(ctx, id)
	content := ""
	if err == nil && vd != nil {
		content = string(vd.Content)
	}

	v2JSON(w, toV2VariableRes(v, content))
}

type v2VariableCreateReq struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

func v2CreateVariable(w http.ResponseWriter, r *http.Request) {
	var req v2VariableCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		v2Error(w, "invalid request body", 400)
		return
	}
	if req.ID == "" {
		v2Error(w, "id は必須です", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	existing, err := dao.SelectVariable(ctx, req.ID)
	if err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	if existing != nil {
		v2Error(w, fmt.Sprintf("変数 '%s' は既に存在します", req.ID), 409)
		return
	}

	vs := &datastore.VariableSet{
		ID:           req.ID,
		Variable:     &datastore.Variable{},
		VariableData: &datastore.VariableData{Content: []byte(req.Content)},
	}
	if err := dao.PutVariable(ctx, vs); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	saved, err := dao.SelectVariable(ctx, req.ID)
	if err != nil || saved == nil {
		v2JSON(w, map[string]string{"status": "created"})
		return
	}
	v2JSON(w, toV2VariableRes(saved, req.Content))
}

type v2VariableUpdateReq struct {
	Content string `json:"content"`
	Version int    `json:"version"`
}

func v2UpdateVariable(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]

	var req v2VariableUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		v2Error(w, "invalid request body", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	existing, err := dao.SelectVariable(ctx, id)
	if err != nil {
		v2Error(w, err.Error(), 500)
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
		v2Error(w, err.Error(), 500)
		return
	}

	updated, err := dao.SelectVariable(ctx, id)
	if err != nil || updated == nil {
		v2JSON(w, map[string]string{"status": "saved"})
		return
	}
	v2JSON(w, toV2VariableRes(updated, req.Content))
}

func v2DeleteVariable(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	if err := dao.RemoveVariable(ctx, id); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, map[string]string{"status": "deleted"})
}
