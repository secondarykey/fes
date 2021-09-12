package internal

import (
	"app/datastore"

	"errors"
	"net/http"
	"strconv"

	"golang.org/x/xerrors"
)

func CreateFormPage(r *http.Request, id string) (*datastore.Page, *datastore.PageData, error) {

	ctx := r.Context()
	p, err := datastore.SelectPage(ctx, id, -1)
	if err != nil {
		return nil, nil, xerrors.Errorf("SelectPage() error: %w", err)
	}

	if p == nil {
		p = &datastore.Page{}
	}

	ver := r.FormValue("version")
	p.TargetVersion = ver

	p.Name = r.FormValue("pageName")
	p.Parent = r.FormValue("parentID")
	seqBuf := r.FormValue("seq")
	v, err := strconv.Atoi(seqBuf)
	if err != nil {
		return nil, nil, xerrors.Errorf("Seq parse error: %w", err)
	}

	p.Seq = v
	p.Description = r.FormValue("pageDescription")
	p.SiteTemplate = r.FormValue("siteTemplateID")
	p.PageTemplate = r.FormValue("pageTemplateID")

	paging, err := strconv.Atoi(r.FormValue("paging"))
	if err == nil {
		p.Paging = paging
	} else {
		//TODO error
	}

	flag := r.FormValue("publish")
	if flag == "on" {
		p.Deleted = false
	} else {
		p.Deleted = true
	}

	if p.SiteTemplate == "" || p.PageTemplate == "" {
		//ページは選択しなくても表示はできるのでOK
		return nil, nil, errors.New("Error:Select Site Template")
	}

	var pd datastore.PageData
	pd.Content = []byte(r.FormValue("pageContent"))

	return p, &pd, nil
}
