package manage

import (
	"app/datastore"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func registerV2SiteAPI(r *mux.Router) {
	r.HandleFunc("/site/", v2GetSite).Methods("GET")
	r.HandleFunc("/site/", v2UpdateSite).Methods("POST")
}

type v2SiteRes struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Root        string    `json:"root"`
	ManageURL   string    `json:"manageURL"`
	Managers    []string  `json:"managers"`
	Version     int       `json:"version"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func toV2SiteRes(s *datastore.Site) v2SiteRes {
	managers := s.Managers
	if managers == nil {
		managers = []string{}
	}
	return v2SiteRes{
		Name:        s.Name,
		Description: s.Description,
		Root:        s.Root,
		ManageURL:   s.ManageURL,
		Managers:    managers,
		Version:     s.Version,
		UpdatedAt:   s.UpdatedAt,
	}
}

func v2GetSite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	site, err := dao.SelectSite(ctx, -1)
	if err != nil {
		if errors.Is(err, datastore.SiteNotFoundError) {
			v2JSON(w, v2SiteRes{Managers: []string{}})
			return
		}
		v2Error(w, err.Error(), 500)
		return
	}
	v2JSON(w, toV2SiteRes(site))
}

type v2SiteUpdateReq struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Root        string   `json:"root"`
	ManageURL   string   `json:"manageURL"`
	Managers    []string `json:"managers"`
	Version     int      `json:"version"`
}

func v2UpdateSite(w http.ResponseWriter, r *http.Request) {
	var req v2SiteUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		v2Error(w, "invalid request body", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	site, err := dao.SelectSite(ctx, -1)
	if err != nil && !errors.Is(err, datastore.SiteNotFoundError) {
		v2Error(w, err.Error(), 500)
		return
	}
	if site == nil {
		site = &datastore.Site{}
	}

	site.SetTargetVersion(fmt.Sprintf("%d", req.Version))
	site.Name = req.Name
	site.Description = req.Description
	site.Root = req.Root
	site.ManageURL = req.ManageURL

	managers := req.Managers
	if managers == nil {
		managers = []string{}
	}
	filtered := make([]string, 0, len(managers))
	for _, m := range managers {
		m = strings.TrimSpace(m)
		if m != "" {
			filtered = append(filtered, m)
		}
	}
	site.Managers = filtered

	if err := dao.PutSite(ctx, site); err != nil {
		v2Error(w, err.Error(), 500)
		return
	}

	updated, err := dao.SelectSite(ctx, -1)
	if err != nil || updated == nil {
		v2JSON(w, map[string]string{"status": "saved"})
		return
	}
	v2JSON(w, toV2SiteRes(updated))
}
