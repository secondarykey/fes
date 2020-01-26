package datastore

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	kerr "github.com/knightso/base/errors"
	"github.com/knightso/base/gae/ds"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
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
	ds.Meta
}

func createSiteKey(r *http.Request) *datastore.Key {
	c := appengine.NewContext(r)
	return datastore.NewKey(c, KindSiteName, "fixing", 0, nil)
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
		page.SetKey(CreatePageKey(r, uid.String()))
	}

	c := appengine.NewContext(r)
	option := &datastore.TransactionOptions{XG: true}
	return datastore.RunInTransaction(c, func(ctx context.Context) error {

		if page != nil {
			err := ds.Put(c, page)
			if err != nil {
				return err
			}
			site.Root = page.Key.StringID()
		}

		key := createSiteKey(r)
		site.SetKey(key)
		err := ds.Put(c, site)
		if err != nil {
			return err
		}
		cacheSite = site
		setDatastoreCache(cacheSite)

		return nil
	}, option)
}

var cacheSite *Site

func SelectSite(r *http.Request, version int) (*Site, error) {

	//バージョン指定がない場合
	if version < 0 {
		if cacheSite != nil {
			return cacheSite, nil
		}
	}

	c := appengine.NewContext(r)
	key := createSiteKey(r)

	var site Site
	var err error
	if version >= 0 {
		err = ds.GetWithVersion(c, key, version, &site)
	} else {
		err = ds.Get(c, key, &site)
	}

	if err != nil {
		if kerr.Root(err) == datastore.ErrNoSuchEntity {
			return nil, SiteNotFoundError
		} else {
			return nil, err
		}
	}
	cacheSite = &site
	setDatastoreCache(cacheSite)
	return &site, nil
}

func setDatastoreCache(site *Site) {
	ds.CacheKinds[KindHTMLName] = site.HTMLCache
	ds.CacheKinds[KindTemplateName] = site.TemplateCache
	ds.CacheKinds[KindTemplateDataName] = site.TemplateCache
	ds.CacheKinds[KindFileName] = site.FileCache
	ds.CacheKinds[KindFileDataName] = site.FileCache
	ds.CacheKinds[KindPageName] = site.PageCache
	ds.CacheKinds[KindPageDataName] = site.PageCache
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

		key1 := page.Key.StringID()
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
