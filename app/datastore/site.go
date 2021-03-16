package datastore

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/xerrors"

	"cloud.google.com/go/datastore"
)

const KindSiteName = "Site"

var (
	SiteNotFoundError = fmt.Errorf("site not found")
)

type Site struct {
	Name          string
	Description   string
	Root          string
	Managers      []string
	HTMLCache     bool
	TemplateCache bool
	FileCache     bool
	PageCache     bool

	TargetVersion string `datastore:"-"`
	Meta
}

func (s *Site) Load(props []datastore.Property) error {
	return datastore.LoadStruct(s, props)
}

func (s *Site) Save() ([]datastore.Property, error) {
	s.update(s.TargetVersion)
	return datastore.SaveStruct(s)
}

func createSiteKey() *datastore.Key {
	return datastore.NameKey(KindSiteName, "fixing", nil)
}

func PutSite(r *http.Request) error {

	ver := r.FormValue("version")
	version, err := strconv.Atoi(ver)
	if err != nil {
		return err
	}

	site, foundErr := SelectSite(r, version)
	if foundErr != nil {
		if foundErr != SiteNotFoundError {
			return foundErr
		}
		site = &Site{}
	}

	site.Name = r.FormValue("name")
	site.Description = r.FormValue("description")
	site.Root = r.FormValue("rootPage")
	site.Managers = strings.Split(r.FormValue("manager"), ",")

	if cache := r.FormValue("htmlCache"); cache != "" {
		if val, err := strconv.ParseBool(cache); err == nil {
			site.HTMLCache = val
		}
	}
	if cache := r.FormValue("templateCache"); cache != "" {
		if val, err := strconv.ParseBool(cache); err == nil {
			site.TemplateCache = val
		}
	}
	if cache := r.FormValue("pageCache"); cache != "" {
		if val, err := strconv.ParseBool(cache); err == nil {
			site.PageCache = val
		}
	}
	if cache := r.FormValue("fileCache"); cache != "" {
		if val, err := strconv.ParseBool(cache); err == nil {
			site.FileCache = val
		}
	}

	var page *Page
	if foundErr != nil {
		page = &Page{
			Name:   "最初のページ",
			Parent: "",
		}
		page.Deleted = true
		uid := uuid.NewV4()
		page.LoadKey(CreatePageKey(uid.String()))
	}

	ctx := r.Context()
	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		if page != nil {
			_, err := tx.Put(page.GetKey(), page)
			if err != nil {
				return xerrors.Errorf("page put error: %w", err)
			}
			site.Root = page.GetKey().Name
		}

		key := createSiteKey()
		site.LoadKey(key)
		_, err := tx.Put(key, site)
		if err != nil {
			return xerrors.Errorf("site put error: %w", err)
		}
		cacheSite = site
		setDatastoreCache(cacheSite)

		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}

	return nil
}

var cacheSite *Site

func SelectSite(r *http.Request, version int) (*Site, error) {

	//バージョン指定がない場合
	if version < 0 {
		if cacheSite != nil {
			return cacheSite, nil
		}
	}

	ctx := r.Context()
	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	key := createSiteKey()

	var site Site
	err = cli.Get(ctx, key, &site)

	if err != nil {
		if errors.Is(err, datastore.ErrNoSuchEntity) {
			return nil, SiteNotFoundError
		} else {
			return nil, err
		}
	}

	//TODO 確認
	if version != 0 {
		site.TargetVersion = fmt.Sprintf("%d", version)
	}

	cacheSite = &site
	setDatastoreCache(cacheSite)
	return &site, nil
}

func setDatastoreCache(site *Site) {
	return
}

type URL struct {
	URL          string
	LastModified string
	Priority     string
	Change       string
	Image        string
	Caption      string
}

func GenerateSitemap(w http.ResponseWriter, r *http.Request) error {

	w.Header().Set("Content-Type", "text/xml")

	scheme := r.URL.Scheme
	if scheme == "" {
		scheme = "https"
	}
	root := fmt.Sprintf("%s://%s/", scheme, r.Host)

	//Page全体でアクセス
	pages, err := SelectPages(r)
	if err != nil {
		return err
	}
	site, err := SelectSite(r, -1)
	if err != nil {
		return err
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

	//Topと同じだった場合
	tmpl, err := template.ParseFiles("templates/map.tmpl")
	if err != nil {
		return err
	}

	err = tmpl.Execute(w, dto)
	if err != nil {
		return err
	}
	return nil
}
