package datastore

import (
	"app/api"
	"context"
	"log"

	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
	"google.golang.org/api/iterator"

	"cloud.google.com/go/datastore"
)

var (
	RootPageNotFoundError = fmt.Errorf("site root not set")
)

const KindPageName = "Page"

type Page struct {
	Name        string
	Seq         int
	Description string
	Parent      string
	Publish     time.Time

	Paging       int
	SiteTemplate string
	PageTemplate string

	TargetVersion string `datastore:"-"`
	Meta
}

func (p *Page) Load(props []datastore.Property) error {
	return datastore.LoadStruct(p, props)
}

func (p *Page) Save() ([]datastore.Property, error) {
	p.update(p.TargetVersion)
	return datastore.SaveStruct(p)
}

func CreatePageKey(id string) *datastore.Key {
	return datastore.NameKey(KindPageName, id, nil)
}

func SelectPages(r *http.Request) ([]Page, error) {

	ctx := r.Context()
	var pages []Page

	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	q := datastore.NewQuery(KindPageName).Filter("Deleted=", false)
	t := cli.Run(ctx, q)
	for {
		var page Page
		_, err := t.Next(&page)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}
	return pages, nil
}

func SelectChildPages(r *http.Request, id string, page int, limit int, mng bool) ([]Page, error) {

	ctx := r.Context()
	var pages []Page

	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	q := datastore.NewQuery(KindPageName).Filter("Parent=", id).Order("Seq").Order("- CreatedAt")
	if !mng {
		q = q.Filter("Deleted=", false)
	}

	//取得件数
	if limit > 0 {
		//カーソルを作成
		q = q.Limit(limit)
	}

	//TODO Cursor
	//ページ数
	if page > 1 {
		//curKey := getChildrenCursorKey(id, page)
	}

	t := cli.Run(ctx, q)
	for {
		var page Page
		_, err := t.Next(&page)
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}

	n := page + 1

	if n > 1 {
		//TODO Cursor
		cur, err := t.Cursor()
		if err != nil {
			return pages, nil
		}
		log.Println("cursor set not implemented", cur)
	}

	return pages, nil
}

func getChildrenCursorKey(id string, p int) string {
	return fmt.Sprintf("children_%s_%d_cursor", id, p)
}

func SelectRootPage(r *http.Request) (*Page, error) {
	site, err := SelectSite(r, -1)
	if err != nil {
		return nil, err
	}
	return SelectPage(r, site.Root, -1)
}

func SelectPage(r *http.Request, id string, version int) (*Page, error) {

	page := Page{}
	ctx := r.Context()
	key := CreatePageKey(id)

	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	if version >= 0 {
		//TODO 調べる
		page.TargetVersion = fmt.Sprintf("%d", version)
	}
	err = cli.Get(ctx, key, &page)

	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
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

	ctx := r.Context()
	ver := r.FormValue("version")

	version, err := strconv.Atoi(ver)
	if err != nil {
		return err
	}

	page, err := SelectPage(r, id, version)
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

	paging, err := strconv.Atoi(r.FormValue("paging"))
	//TODO err?
	if err == nil {
		page.Paging = paging
	}

	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		page.LoadKey(CreatePageKey(id))
		_, err = tx.Put(page.GetKey(), page)
		if err != nil {
			return err
		}
		pageData.LoadKey(CreatePageDataKey(id))
		_, err = tx.Put(pageData.GetKey(), pageData)
		if err != nil {
			return err
		}

		err = SaveFile(r, id, api.FileTypePageImage)
		if err != nil {
			//ファイル指定なしの場合の動作
		}

		//TODO Deletedにされている場合、HTMLを検索して削除

		return nil
	})
	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}

	return nil
}

func UsingTemplate(r *http.Request, id string) bool {

	var err error
	ctx := r.Context()

	cli, err := createClient(ctx)
	if err != nil {
		return true
		//return xerrors.Errorf("createClient() error: %w", err)
	}

	siteQ := datastore.NewQuery(KindPageName).Filter("SiteTemplate=", id).Limit(1)
	siteT := cli.Run(ctx, siteQ)

	var page Page
	_, err = siteT.Next(&page)
	if errors.Is(err, iterator.Done) {
		pageQ := datastore.NewQuery(KindPageName).Filter("PageTemplate=", id).Limit(1)
		pageT := cli.Run(ctx, pageQ)
		_, err = pageT.Next(&page)
		if errors.Is(err, iterator.Done) {
			return false
		}
	}
	return true
}

func RemovePage(r *http.Request, id string) error {

	var err error
	ctx := r.Context()

	children, err := SelectChildPages(r, id, 0, 0, false)
	if err != nil {
		return fmt.Errorf("Datastore Error SelectChildPages child page[%v]", err)
	}

	if children != nil {
		return fmt.Errorf("Exist child page[%s]", id)
	}

	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		pkey := CreatePageKey(id)
		err = tx.Delete(pkey)
		if err != nil {
			return err
		}
		pdkey := CreatePageDataKey(id)
		err = tx.Delete(pdkey)
		if err != nil {
			return err
		}
		if ExistFile(r, id) {
			return RemoveFile(r, id)
		}
		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}
	return nil
}

func PutPageSequence(r *http.Request, ids string, enables string, verCsv string) error {

	idArray := strings.Split(ids, ",")
	enableArray := strings.Split(enables, ",")
	versionsArray := strings.Split(verCsv, ",")

	keys := make([]*datastore.Key, len(idArray))
	deleteds := make([]bool, len(enableArray))
	versions := make([]string, len(versionsArray))

	for idx, id := range idArray {
		key := CreatePageKey(id)
		keys[idx] = key

		flagBuf := enableArray[idx]
		flag, err := strconv.ParseBool(flagBuf)
		if err != nil {
			return err
		}
		deleteds[idx] = !flag

		verBuf := versionsArray[idx]
		versions[idx] = verBuf
	}

	ctx := r.Context()
	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	pages := make([]*Page, len(keys))

	err = cli.GetMulti(ctx, keys, pages)
	if err != nil {
		return err
	}

	// TODO これでいいか確認
	for idx, page := range pages {
		page.TargetVersion = versions[idx]
	}

	for idx, page := range pages {
		page.Seq = idx + 1
		page.Deleted = deleteds[idx]
	}
	_, err = cli.PutMulti(ctx, keys, pages)
	if err != nil {
		return xerrors.Errorf("page PutMulti() error: %w", err)
	}
	return nil
}

func SelectReferencePages(r *http.Request, id string, typ string) ([]Page, error) {

	ctx := r.Context()
	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	var pages []Page
	filter := "SiteTemplate="
	if typ == "2" {
		filter = "PageTemplate="
	}

	q := datastore.NewQuery(KindPageName).Filter(filter, id)
	t := cli.Run(ctx, q)
	for {
		var page Page
		_, err := t.Next(&page)
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}
	return pages, nil
}

const KindPageDataName = "PageData"

type PageData struct {
	Key     *datastore.Key `datastore:"__key__"`
	Content []byte         `datastore:",noindex"`
}

func (d *PageData) GetKey() *datastore.Key {
	return d.Key
}

func (d *PageData) LoadKey(k *datastore.Key) error {
	d.Key = k
	return nil
}

func CreatePageDataKey(id string) *datastore.Key {
	return datastore.NameKey(KindPageDataName, id, nil)
}

func SelectPageData(r *http.Request, id string) (*PageData, error) {

	page := PageData{}
	ctx := r.Context()
	key := CreatePageDataKey(id)

	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	err = cli.Get(ctx, key, &page)
	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
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

func PageTree(ctx context.Context) (*Tree, error) {

	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	var pages []*Page
	q := datastore.NewQuery(KindPageName)

	keys, err := cli.GetAll(ctx, q, &pages)
	if err != nil {
		return nil, err
	}

	parentMap := make(map[string][]*Page)

	//キーマップの作成
	for idx, elm := range pages {
		key := keys[idx]
		elm.LoadKey(key)
		children, ok := parentMap[elm.Parent]
		if !ok {
			children = make([]*Page, 0)
		}
		children = append(children, elm)
		parentMap[elm.Parent] = children
	}

	//全データのソート
	for _, slice := range parentMap {
		sort.Slice(slice, func(i, j int) bool {
			pageI := slice[i]
			pageJ := slice[j]
			if pageI.Seq < pageJ.Seq {
				return true
			} else if pageI.Seq > pageJ.Seq {
				return false
			}
			return pageI.CreatedAt.Unix() > pageJ.CreatedAt.Unix()
		})
	}

	roots := parentMap[""]
	tree := createTree(1, roots[0], roots[0].Key.Name, parentMap)

	return tree, nil
}

func createTree(indent int, page *Page, key string, parentMap map[string][]*Page) *Tree {

	tree := Tree{
		Page:     page,
		Children: make([]*Tree, 0),
		Indent:   indent,
	}

	children, ok := parentMap[key]
	if ok {
		for _, child := range children {
			childTree := createTree(indent+1, child, child.Key.Name, parentMap)
			tree.Children = append(tree.Children, childTree)
		}
	}
	return &tree
}
