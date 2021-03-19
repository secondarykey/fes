package internal

import (
	"app/datastore"
	"context"
	"html/template"
	"net/http"
	"time"

	"golang.org/x/xerrors"
)

type URL struct {
	URL          string
	LastModified string
	Priority     string
	Change       string
	Image        string
	Caption      string
}

func GenerateSitemap(ctx context.Context, root string, w http.ResponseWriter) error {

	//Page全体でアクセス
	pages, err := datastore.SelectPages(ctx)
	if err != nil {
		return xerrors.Errorf("datastore.SelectPages() error: %w", err)
	}
	site, err := datastore.SelectSite(ctx, -1)
	if err != nil {
		return xerrors.Errorf("datastore.SelectSite() error: %w", err)
	}

	rootId := site.Root

	urls := make([]URL, len(pages))
	//Page数回繰り返す
	for idx, page := range pages {

		key1 := page.Key.Name
		key2 := "page/" + key1
		if key1 == rootId {
			key2 = ""
		}

		url := URL{}
		url.URL = root + key2
		url.LastModified = page.UpdatedAt.Format(time.RFC3339)
		url.Change = "weekly"
		url.Priority = "0.8"
		url.Image = root + "file/" + key1
		url.Caption = page.Description

		urls[idx] = url
	}

	dto := struct {
		Header template.HTML
		Pages  []URL
	}{template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`), urls}

	w.Header().Set("Content-Type", "text/xml")
	err = WriteTemplate(w, dto, "map.tmpl")
	if err != nil {
		return xerrors.Errorf("WriteTemplate() error: %w", err)
	}

	return nil
}
