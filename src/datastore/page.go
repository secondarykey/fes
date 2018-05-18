package datastore

import (
	"net/http"

	"github.com/gorilla/mux"

	kerr "github.com/knightso/base/errors"
	"github.com/knightso/base/gae/ds"
	"golang.org/x/net/context"

	"errors"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"fmt"
)

const KIND_PAGE = "Page"

type Page struct {
	Name        string
	Seq         int
	Description string
	Parent      string

	SiteTemplate string
	PageTemplate string
	ds.Meta
}

func CreatePageKey(r *http.Request, id string) *datastore.Key {
	c := appengine.NewContext(r)
	return datastore.NewKey(c, KIND_PAGE, id, 0, nil)
}

func SelectChildPages(r *http.Request, id string) ([]Page, error) {
	c := appengine.NewContext(r)
	var pages []Page
	q := datastore.NewQuery(KIND_PAGE).Filter("Parent=", id).Order("Seq").Order("CreatedAt")
	t := q.Run(c)
	for {
		var page Page
		key, err := t.Next(&page)
		if err == datastore.Done {
			break
		}

		if err != nil {
			return nil, err
		}
		page.SetKey(key)
		pages = append(pages, page)
	}
	return pages, nil
}

func SelectRootPage(r *http.Request) (*Page, error) {
	site := GetSite(r)
	if site.Root == "" {
		return nil, nil
	}
	return SelectPage(r, site.Root)
}

func SelectPage(r *http.Request, id string) (*Page, error) {
	page := Page{}
	c := appengine.NewContext(r)
	key := datastore.NewKey(c, KIND_PAGE, id, 0, nil)
	err := ds.Get(c, key, &page)
	if err != nil {
		if kerr.Root(err) != datastore.ErrNoSuchEntity {
			return nil, err
		} else {
			return nil, nil
		}
	}
	return &page, nil
}

func PutPage(r *http.Request) error {

	var err error

	vars := mux.Vars(r)
	id := vars["key"]

	c := appengine.NewContext(r)
	page, err := SelectPage(r, id)
	if err != nil {
		return err
	}

	if page == nil {
		page = &Page{}
	}
	page.Name = r.FormValue("pageName")
	page.Parent = r.FormValue("parentID")
	page.Description = r.FormValue("pageDescription")
	page.SiteTemplate = r.FormValue("siteTemplateID")
	page.PageTemplate = r.FormValue("pageTemplateID")

	if page.SiteTemplate == "" {
		//ページは選択しなくても表示はできるのでOK
		return errors.New("Error:Select Site Template")
	}

	//Data については検索せずに更新
	pageData := &PageData{
		Content: datastore.ByteString(r.FormValue("pageContent")),
	}

	option := &datastore.TransactionOptions{XG: true}
	return datastore.RunInTransaction(c, func(ctx context.Context) error {
		page.SetKey(CreatePageKey(r, id))
		err = ds.Put(c, page)
		if err != nil {
			return err
		}
		pageData.SetKey(CreatePageDataKey(r, id))
		err = ds.Put(c, pageData)
		if err != nil {
			return err
		}

		err = SaveFile(r, id)
		if err != nil {
			//ファイル指定なしの場合の動作
		}

		//一番親ページの場合
		if page.Parent == "" {
			//SiteのページKeyを変更する
			err = SetRoot(r, page.Key.StringID())
			if err != nil {
				return err
			}
		}
		return nil
	}, option)
}

func RemovePage(r *http.Request, id string) error {

	var err error
	c := appengine.NewContext(r)

	children,err := SelectChildPages(r,id)

	if  children != nil {

		return fmt.Errorf("Exist child page[%s]",id)
	}

	option := &datastore.TransactionOptions{XG: true}
	return datastore.RunInTransaction(c, func(ctx context.Context) error {
		pkey := CreatePageKey(r, id)
		err = ds.Delete(c, pkey)
		if err != nil {
			return nil
		}
		pdkey := CreatePageDataKey(r, id)
		err = ds.Delete(c, pdkey)
		if err != nil {
			return nil
		}
		if ExistFile(r,id) {
			return RemoveFile(r, id)
		}
		return nil
	}, option)
}

const KIND_PAGEDATA = "PageData"

type PageData struct {
	key     *datastore.Key
	Content []byte
}

func (d *PageData) GetKey() *datastore.Key {
	return d.key
}

func (d *PageData) SetKey(k *datastore.Key) {
	d.key = k
}

func CreatePageDataKey(r *http.Request, id string) *datastore.Key {
	c := appengine.NewContext(r)
	return datastore.NewKey(c, KIND_PAGEDATA, id, 0, nil)
}

func SelectPageData(r *http.Request, id string) (*PageData, error) {

	page := PageData{}
	c := appengine.NewContext(r)
	key := datastore.NewKey(c, KIND_PAGEDATA, id, 0, nil)
	err := ds.Get(c, key, &page)
	if err != nil {
		if kerr.Root(err) != datastore.ErrNoSuchEntity {
			return nil, err
		} else {
			return nil, nil
		}
	}
	return &page, nil
}
