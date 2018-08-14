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

