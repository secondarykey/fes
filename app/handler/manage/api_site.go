package manage

import (
	"app/datastore"
	"app/handler/internal"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func registerSiteAPI(r *mux.Router) {
	r.HandleFunc("/site/", apiGetSite).Methods("GET")
	r.HandleFunc("/site/", apiUpdateSite).Methods("POST")
}

type apiSiteRes struct {
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Root              string    `json:"root"`
	ManageURL         string    `json:"manageURL"`
	BaseURL           string    `json:"baseURL"`
	Managers          []string  `json:"managers"`
	ArchiveBucket     string    `json:"archiveBucket"`
	ArchiveNames      []string  `json:"archiveNames"`
	ArchiveDailyLimit int       `json:"archiveDailyLimit"`
	Version           int       `json:"version"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func toAPISiteRes(s *datastore.Site) apiSiteRes {
	managers := s.Managers
	if managers == nil {
		managers = []string{}
	}
	archiveNames := s.ArchiveNames
	if archiveNames == nil {
		archiveNames = []string{}
	}
	return apiSiteRes{
		Name:              s.Name,
		Description:       s.Description,
		Root:              s.Root,
		ManageURL:         s.ManageURL,
		BaseURL:           s.BaseURL,
		Managers:          managers,
		ArchiveBucket:     s.ArchiveBucket,
		ArchiveNames:      archiveNames,
		ArchiveDailyLimit: s.ArchiveDailyLimit,
		Version:           s.Version,
		UpdatedAt:         s.UpdatedAt,
	}
}

func apiGetSite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	site, err := dao.SelectSite(ctx, -1)
	if err != nil {
		if errors.Is(err, datastore.SiteNotFoundError) {
			apiJSON(w, apiSiteRes{Managers: []string{}, ArchiveNames: []string{}})
			return
		}
		apiError(w, err.Error(), 500)
		return
	}
	apiJSON(w, toAPISiteRes(site))
}

type apiSiteUpdateReq struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Root              string   `json:"root"`
	ManageURL         string   `json:"manageURL"`
	BaseURL           string   `json:"baseURL"`
	Managers          []string `json:"managers"`
	ArchiveBucket     string   `json:"archiveBucket"`
	ArchiveNames      []string `json:"archiveNames"`
	ArchiveDailyLimit int      `json:"archiveDailyLimit"`
	Version           int      `json:"version"`
}

func apiUpdateSite(w http.ResponseWriter, r *http.Request) {
	var req apiSiteUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, "invalid request body", 400)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	site, err := dao.SelectSite(ctx, -1)
	if err != nil && !errors.Is(err, datastore.SiteNotFoundError) {
		apiError(w, err.Error(), 500)
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
	site.BaseURL = strings.TrimSpace(req.BaseURL)

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

	site.ArchiveBucket = strings.TrimSpace(req.ArchiveBucket)

	archiveNames := req.ArchiveNames
	if archiveNames == nil {
		archiveNames = []string{}
	}
	filteredNames := make([]string, 0, len(archiveNames))
	for _, n := range archiveNames {
		n = strings.TrimSpace(n)
		if n != "" {
			filteredNames = append(filteredNames, n)
		}
	}
	site.ArchiveNames = filteredNames
	site.ArchiveDailyLimit = req.ArchiveDailyLimit

	if err := dao.PutSite(ctx, site); err != nil {
		apiError(w, err.Error(), 500)
		return
	}

	// GCS アーカイブルーターを動的に更新
	if internal.GCSArchiveRouter != nil {
		internal.GCSArchiveRouter.Update(site.ArchiveBucket, site.ArchiveNames)
		if site.ArchiveDailyLimit > 0 {
			internal.GCSArchiveRouter.SetLimit(int64(site.ArchiveDailyLimit))
		} else {
			internal.GCSArchiveRouter.SetLimit(internal.DefaultArchiveDailyLimit)
		}
	}

	updated, err := dao.SelectSite(ctx, -1)
	if err != nil || updated == nil {
		apiJSON(w, map[string]string{"status": "saved"})
		return
	}
	apiJSON(w, toAPISiteRes(updated))
}
