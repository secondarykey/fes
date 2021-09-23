package internal

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"app/datastore"

	"golang.org/x/xerrors"
)

func CreateFormSite(r *http.Request) (*datastore.Site, error) {

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	ver := r.FormValue("version")
	version, err := strconv.Atoi(ver)
	if err != nil {
		return nil, xerrors.Errorf("site version error: %w", err)
	}

	site, err := dao.SelectSite(ctx, version)
	if !errors.Is(err, datastore.SiteNotFoundError) {
		return nil, xerrors.Errorf("SelectSite() error: %w", err)
	} else {
		site = &datastore.Site{}
	}

	site.Name = r.FormValue("name")
	site.Description = r.FormValue("description")
	site.Root = r.FormValue("rootPage")
	site.Managers = strings.Split(r.FormValue("manager"), ",")

	return site, nil
}
