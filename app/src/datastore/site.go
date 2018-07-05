package datastore

import (
	"net/http"

	"github.com/knightso/base/gae/ds"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"log"
	"strconv"
)

const KIND_SITE = "Site"

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

func SetRoot(r *http.Request, id string) error {
	c := appengine.NewContext(r)

	site := GetSite(r)
	site.Root = id
	err := ds.Put(c, site)
	if err != nil {
		return err
	}
	gSite = site
	return nil
}

var gSite *Site

func PutSite(r *http.Request) error {

	gSite = GetSite(r)

	gSite.Name = r.FormValue("name")
	gSite.Description = r.FormValue("description")
	gSite.Root = r.FormValue("rootPage")

	if cache := r.FormValue("htmlCache") ; cache != "" {
		if val,err := strconv.ParseBool(cache) ; err == nil {
			gSite.HTMLCache = val
		}
	}
	if cache := r.FormValue("templateCache") ; cache != "" {
		if val,err := strconv.ParseBool(cache) ; err == nil {
			gSite.TemplateCache = val
		}
	}
	if cache := r.FormValue("pageCache") ; cache != "" {
		if val,err := strconv.ParseBool(cache) ; err == nil {
			gSite.PageCache = val
		}
	}
	if cache := r.FormValue("fileCache") ; cache != "" {
		if val,err := strconv.ParseBool(cache) ; err == nil {
			gSite.FileCache = val
		}
	}

	c := appengine.NewContext(r)
	key := createSiteKey(r)
	gSite.SetKey(key)
	err := ds.Put(c, gSite)
	if err != nil {
		return err
	}

	return nil
}

func GetSite(r *http.Request) *Site {

	if gSite != nil {
		return gSite
	}

	site := Site{
		Name:        "サイト名",
		Description: "サイトの説明",
		TemplateCache:true,
		PageCache:true,
		FileCache:true,
	}

	c := appengine.NewContext(r)
	key := createSiteKey(r)
	err := ds.Get(c, key, &site)
	if err != nil {
		log.Println("サイトデータ取得失敗")
	}

	//キャッシュ設定
	ds.CacheKinds[KIND_TEMPLATE] = site.TemplateCache
	ds.CacheKinds[KIND_TEMPLATEDATA] = site.TemplateCache
	ds.CacheKinds[KIND_PAGE] = site.PageCache
	ds.CacheKinds[KIND_PAGEDATA] = site.PageCache
	ds.CacheKinds[KIND_FILE] = site.FileCache
	ds.CacheKinds[KIND_FILEDATA] = site.FileCache

	if site.Key == nil {
		site.SetKey(key)
		err = ds.Put(c, &site)
		if err != nil {
			log.Println("サイトデータPut失敗")
		}
	}

	gSite = &site
	return gSite
}
