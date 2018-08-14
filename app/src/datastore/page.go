package datastore

import (
	"net/http"
	"fmt"
	"api"
	"time"
	"strings"
	"strconv"
	"errors"

	"github.com/gorilla/mux"

	kerr "github.com/knightso/base/errors"
	"github.com/knightso/base/gae/ds"
	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
	"log"
	"sort"
)

var (
	RootPageNotFoundError = fmt.Errorf("site root not set")
)

const KIND_PAGE = "Page"

type Page struct {
	Name        string
	Seq         int
	Description string
	Parent      string
	Publish     time.Time

	Paging       int
	SiteTemplate string
	PageTemplate string
	ds.Meta
}

func CreatePageKey(r *http.Request, id string) *datastore.Key {
	c := appengine.NewContext(r)
	return datastore.NewKey(c, KIND_PAGE, id, 0, nil)
}

func SelectPages(r *http.Request) ([]Page, error) {
	c := appengine.NewContext(r)
	var pages []Page
	q := datastore.NewQuery(KIND_PAGE).Filter("Deleted=",false)
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

func SelectChildPages(r *http.Request, id string,page int,limit int,mng bool) ([]Page, error) {

	c := appengine.NewContext(r)
	var pages []Page

	q := datastore.NewQuery(KIND_PAGE).Filter("Parent=", id).Order("Seq").Order("- CreatedAt")
	if !mng {
		q = q.Filter("Deleted=",false)
	}

	//取得件数
	if limit > 0 {
		//カーソルを作成
		q = q.Limit(limit)
	}

	//ページ数
	if page > 1 {
		curKey := getChildrenCursorKey(id,page)
		item, err := memcache.Get(c, getChildrenCursorKey(id,page))
		if err == nil {
			cursor := string(item.Value)
			log.Printf("Key:%s, Get Cursor:%s",curKey,item.Value)

			cur,err := datastore.DecodeCursor(cursor)
			if err == nil {
				q = q.Start(cur)
			}
		}
	}

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

	n := page + 1

	if n > 1 {
		cur, err := t.Cursor()
		if err != nil {
			return pages, nil
		}

		curKey := getChildrenCursorKey(id,n)
		log.Printf("Key:%s, Set Cursor:%s",curKey,cur.String())

		err = memcache.Set(c, &memcache.Item{
			Key:   curKey,
			Value: []byte(cur.String()),
		})
	}

	return pages, nil
}

func getChildrenCursorKey(id string,p int) string {
	return fmt.Sprintf("children_%s_%d_cursor",id,p)
}

func SelectRootPage(r *http.Request) (*Page, error) {
	site,err := SelectSite(r)
	if err != nil {
		return nil,err
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
	page.SetKey(key)
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

	flag := r.FormValue("publish")
	if flag == "on" {
		page.Deleted = false
	} else {
		page.Deleted = true
	}

	if page.SiteTemplate == "" {
		//ページは選択しなくても表示はできるのでOK
		return errors.New("Error:Select Site Template")
	}

	//Data については検索せずに更新
	pageData := &PageData{
		Content: []byte(r.FormValue("pageContent")),
	}

	paging,err := strconv.Atoi(r.FormValue("paging"))
	//TODO err?
	if err == nil {
		page.Paging = paging
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

		err = SaveFile(r, id,api.PAGE_IMAGE)
		if err != nil {
			//ファイル指定なしの場合の動作
		}

		//TODO Deletedにされている場合、HTMLを検索して削除

		return nil
	}, option)
}

func UsingTemplate(r *http.Request, id string) bool {

	var err error
	c := appengine.NewContext(r)
	siteQ := datastore.NewQuery(KIND_PAGE).Filter("SiteTemplate=", id).Limit(1)
	siteT := siteQ.Run(c)
	var page Page
	_, err = siteT.Next(&page)
	if err == datastore.Done {
		pageQ := datastore.NewQuery(KIND_PAGE).Filter("PageTemplate=", id).Limit(1)
		pageT := pageQ.Run(c)
		_, err = pageT.Next(&page)
		if err == datastore.Done {
			return false
		}
	}
	return true
}

func RemovePage(r *http.Request, id string) error {

	var err error
	c := appengine.NewContext(r)

	children,err := SelectChildPages(r,id,0,0,false)
	if  err != nil {
		return fmt.Errorf("Datastore Error SelectChildPages child page[%v]",err)
	}

	if  children != nil {
		return fmt.Errorf("Exist child page[%s]",id)
	}

	option := &datastore.TransactionOptions{XG: true}
	return datastore.RunInTransaction(c, func(ctx context.Context) error {
		pkey := CreatePageKey(r, id)
		err = ds.Delete(c, pkey)
		if err != nil {
			return err
		}
		pdkey := CreatePageDataKey(r, id)
		err = ds.Delete(c, pdkey)
		if err != nil {
			return err
		}
		if ExistFile(r,id) {
			return RemoveFile(r, id)
		}
		return nil
	}, option)
}

func PutPageSequence(r *http.Request, ids string,enables string) (error) {

	idArray := strings.Split(ids,",")
	enableArray := strings.Split(enables,",")

	keys := make([]*datastore.Key,len(idArray))
	deleteds := make([]bool,len(enableArray))

	for idx,id := range idArray {
		key := CreatePageKey(r,id)
		keys[idx] = key

		flagBuf := enableArray[idx]
		flag,err := strconv.ParseBool(flagBuf)
		if err != nil {
			return err
		}
		deleteds[idx] = !flag
	}

	c := appengine.NewContext(r)

	pages := make([]*Page,len(keys))
	err := ds.GetMulti(c,keys,pages)
	if err != nil {
		return err
	}
	for idx,page := range pages {
		page.Seq = idx + 1
		page.Deleted = deleteds[idx]
	}

	return ds.PutMulti(c,keys,pages)
}

func SelectReferencePages(r *http.Request,id string,typ string) ([]Page,error){

	c := appengine.NewContext(r)
	var pages []Page
	filter := "SiteTemplate="
	if typ == "2" {
		filter = "PageTemplate="
	}

	q := datastore.NewQuery(KIND_PAGE).Filter(filter,id)
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
	return pages,nil
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

type Tree struct {
	Page     *Page
	Children []*Tree
	Indent   int
}

func PageTree(r *http.Request) (*Tree,error) {

	c := appengine.NewContext(r)
	var pages []*Page
	q := datastore.NewQuery(KIND_PAGE)

	keys,err := q.GetAll(c,&pages)
	if err != nil {
		return nil,err
	}

	parentMap := make(map[string][]*Page)

	//キーマップの作成
	for idx,elm := range pages {
		key := keys[idx]
		elm.SetKey(key)
		children,ok := parentMap[elm.Parent]
		if !ok {
			children = make([]*Page,0)
		}
		children = append(children,elm)
		parentMap[elm.Parent] = children
	}

	//全データのソート
	for _,slice := range parentMap {
		sort.Slice(slice,func(i,j int) bool {
			pageI := slice[i]
			pageJ := slice[j]
			if pageI.Seq < pageJ.Seq {
				return true
			} else	if pageI.Seq > pageJ.Seq {
				return false
			}
			return pageI.CreatedAt.Unix() > pageJ.CreatedAt.Unix()
		})
	}

	roots := parentMap[""]
	tree := createTree(1,roots[0],roots[0].Key.StringID(),parentMap)

	return tree,nil
}

func createTree(indent int,page *Page,key string,parentMap map[string][]*Page) (*Tree) {

	tree := Tree{
		Page:page,
		Children:make([]*Tree,0),
		Indent:indent,
	}

	children,ok := parentMap[key]
	if ok {
		for _,child := range children {
			childTree := createTree(indent + 1,child,child.Key.StringID(),parentMap)
			tree.Children = append(tree.Children,childTree)
		}
	}
	return &tree
}
