package internal

import (
	"app/datastore"

	"errors"
	"net/http"
	"strconv"
)

func CreateFormPage(r *http.Request) (*datastore.Page, *datastore.PageData, error) {

	var p datastore.Page
	var pd datastore.PageData

	ver := r.FormValue("version")
	p.TargetVersion = ver
	p.Name = r.FormValue("pageName")
	p.Parent = r.FormValue("parentID")
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

	pd.Content = []byte(r.FormValue("pageContent"))

	return &p, &pd, nil
}
