package datastore

import (
	"net/http"

	"github.com/knightso/base/gae/ds"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"log"
)

const KIND_SITE = "Site"

type Site struct {
	Name        string
	Description string
	Root        string
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
	}

	c := appengine.NewContext(r)
	key := createSiteKey(r)
	err := ds.Get(c, key, &site)
	if err != nil {
		log.Println("サイトデータ取得失敗")
	}

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
