package datastore

import (
	"net/http"
	"strconv"
	"fmt"

	"golang.org/x/net/context"
	"github.com/knightso/base/gae/ds"
	kerr "github.com/knightso/base/errors"
	"github.com/satori/go.uuid"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"io"
	"html/template"
	"time"
)

const KIND_SITE = "Site"

var (
	SiteNotFoundError = fmt.Errorf("site not found")
)

type Site struct {
	Name        string
	Description string
	Root        string
	HTMLCache     bool
	TemplateCache bool
	FileCache     bool
	PageCache     bool
	ds.Meta
}

func createSiteKey(r *http.Request) *datastore.Key {
	c := appengine.NewContext(r)
	return datastore.NewKey(c, KIND_SITE, "fixing", 0, nil)
}

func PutSite(r *http.Request) error {

	site,foundErr := SelectSite(r)
	if foundErr != nil {
		if foundErr != SiteNotFoundError {
			return foundErr
		}
		site = &Site{}
	}

	site.Name = r.FormValue("name")
	site.Description = r.FormValue("description")
	site.Root = r.FormValue("rootPage")

	if cache := r.FormValue("htmlCache") ; cache != "" {
		if val,err := strconv.ParseBool(cache) ; err == nil {
			site.HTMLCache = val
		}
	}
	if cache := r.FormValue("templateCache") ; cache != "" {
		if val,err := strconv.ParseBool(cache) ; err == nil {
			site.TemplateCache = val
		}
	}
	if cache := r.FormValue("pageCache") ; cache != "" {
		if val,err := strconv.ParseBool(cache) ; err == nil {
			site.PageCache = val
		}
	}
	if cache := r.FormValue("fileCache") ; cache != "" {
		if val,err := strconv.ParseBool(cache) ; err == nil {
			site.FileCache = val
		}
	}

	var page *Page
	if foundErr != nil {
		page = &Page{
			Name : "最初のページ",
			Parent: "",
		}
		page.Deleted = true
		uid, err := uuid.NewV4()
		if err != nil {
			return err
		}
		page.SetKey(CreatePageKey(r,uid.String()))
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

		return nil
	}, option)
}

var cacheSite *Site
func SelectSite(r *http.Request) (*Site,error) {

	if cacheSite != nil {
		return cacheSite,nil
	}

	c := appengine.NewContext(r)
	key := createSiteKey(r)

	var site Site
	err := ds.Get(c, key, &site)
	if err != nil {
		if kerr.Root(err) == datastore.ErrNoSuchEntity {
			return nil,SiteNotFoundError
		} else {
			return nil, err
		}
	}
	cacheSite = &site
	return &site,nil
}

type URL struct {
	URL string
	LastModified string
	Priority string
	Change string
	Image string
	Caption string
}

func GenerateSitemap(w io.Writer,r *http.Request,root string) error {

	//Page全体でアクセス
	pages,err := SelectPages(r)
	if err != nil {
		return err
	}
	urls := make([]URL,len(pages))
	//Page数回繰り返す
	for idx,page := range pages {

		url := URL{}
		url.URL = root + "page/" + page.Key.StringID()
		url.LastModified = page.UpdatedAt.Format(time.RFC3339)
		url.Change = "weekly"
		url.Priority = "0.8"
		url.Image = root + "file/" + page.Key.StringID()
		url.Caption = page.Description

		urls[idx] = url
	}

	dto := struct {
		Header template.HTML
		Pages []URL
	}{template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),urls}

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
