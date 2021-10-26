package datastore

import (
	"context"

	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/xerrors"
	"google.golang.org/api/iterator"

	"cloud.google.com/go/datastore"
)

const (
	ErrorPageID = "ErrorPage"
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

	Meta
}

func (p *Page) Load(props []datastore.Property) error {
	err := datastore.LoadStruct(p, props)
	if err != nil {
		return xerrors.Errorf("page Load() error: %w", err)
	}
	return nil
}

func (p *Page) Save() ([]datastore.Property, error) {
	err := p.update()
	if err != nil {
		return nil, xerrors.Errorf("Meta update() error: %w", err)
	}
	return datastore.SaveStruct(p)
}

func CreatePageKey(id string) *datastore.Key {
	return datastore.NameKey(KindPageName, id, createSiteKey())
}

func (dao *Dao) SelectPages(ctx context.Context, ids ...string) ([]Page, error) {
	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	keys := make([]*datastore.Key, 0, len(ids))
	for _, id := range ids {
		key := CreatePageKey(id)
		keys = append(keys, key)
	}

	pages := make([]Page, len(keys))
	err = cli.GetMulti(ctx, keys, pages)
	if err != nil {
		return nil, xerrors.Errorf("client.GetAll() error: %w", err)
	}

	return pages, nil
}

func (dao *Dao) SelectAllPages(ctx context.Context) ([]*Page, error) {

	var pages []*Page
	cli, err := dao.createClient(ctx)
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

			return pages, nil
		}
		if err != nil {
			return nil, err
		}
		pages = append(pages, &page)
	}
	return pages, nil
}

func (dao *Dao) SelectChildrenPage(ctx context.Context, id string, cur string, limit int, mng bool) ([]Page, string, error) {

	var pages []Page

	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, "", xerrors.Errorf("createClient() error: %w", err)
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

	//ページ数
	if cur != "" && cur != NoLimitCursor {
		cursor, err := datastore.DecodeCursor(cur)
		if err != nil {
			return nil, "", xerrors.Errorf("datastore.Decode() error: %w", err)
		}
		q = q.Start(cursor)
	}

	t := cli.Run(ctx, q)
	for {
		var page Page
		_, err := t.Next(&page)
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, "", xerrors.Errorf("Page Next() error: %w", err)
		}
		pages = append(pages, page)
	}

	cursor, err := t.Cursor()
	if err != nil {
		return nil, "", xerrors.Errorf("Page Cursor() error: %w", err)
	}

	return pages, cursor.String(), nil
}

func (dao *Dao) SelectRootPage(ctx context.Context) (*Page, error) {
	site, err := dao.SelectSite(ctx, -1)
	if err != nil {
		return nil, xerrors.Errorf("SelectSite() error: %w", err)
	}
	return dao.SelectPage(ctx, site.Root, -1)
}

func (dao *Dao) SelectPage(ctx context.Context, id string, version int) (*Page, error) {

	page := Page{}

	key := CreatePageKey(id)
	cli, err := dao.createClient(ctx)
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

func (dao *Dao) GetErrorPage(ctx context.Context) (*Page, error) {
	p, err := dao.SelectPage(ctx, ErrorPageID, -1)
	if err != nil {
		return nil, xerrors.Errorf("SelectPage() error: %w", err)
	}
	return p, nil
}

type PageSet struct {
	ID       string
	Page     *Page
	PageData *PageData
	FileSet  *FileSet
}

func (dao *Dao) PutPage(ctx context.Context, p *PageSet) error {

	id := p.ID

	p.Page.LoadKey(CreatePageKey(id))
	p.PageData.LoadKey(CreatePageDataKey(id))

	cli, err := dao.createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		_, err = tx.Put(p.Page.GetKey(), p.Page)
		if err != nil {
			return xerrors.Errorf("File Put() error: %w", err)
		}

		_, err = tx.Put(p.PageData.GetKey(), p.PageData)
		if err != nil {
			return xerrors.Errorf("FileData Put() error: %w", err)
		}

		//ファイル指定なしの場合の動作
		if p.FileSet.File != nil {
			err = dao.SaveFile(ctx, p.FileSet)
			if err != nil {
				return xerrors.Errorf("SaveFile() error: %w", err)
			}
		}

		//TODO Deletedにされている場合、HTMLを検索して削除
		//     このタイミングで削除を行うと、公開ページごと削除されてしまう。
		//     公開ページを更新時に削除する仕組みか、テスト中に表示できる仕組みを他で用意する
		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}

	return nil
}

func (dao *Dao) UsingTemplate(ctx context.Context, id string) (bool, error) {

	var err error
	cli, err := dao.createClient(ctx)
	if err != nil {
		return true, xerrors.Errorf("createClient() error: %w", err)
	}

	siteQ := datastore.NewQuery(KindPageName).Filter("SiteTemplate=", id).Limit(1).KeysOnly()
	keys, err := cli.GetAll(ctx, siteQ, nil)
	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
			return true, xerrors.Errorf("SiteTemplate Select error: %w", err)
		}
	}

	if len(keys) > 0 {
		return true, fmt.Errorf("This ID is used as a Site Template PageID=%v", getIDs(keys))
	}

	pageQ := datastore.NewQuery(KindPageName).Filter("PageTemplate=", id).Limit(1).KeysOnly()
	keys, err = cli.GetAll(ctx, pageQ, nil)
	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
			return true, xerrors.Errorf("SiteTemplate Select error: %w", err)
		}
	}

	if len(keys) > 0 {
		return true, fmt.Errorf("This ID is used as a Page Template PageID=%v", getIDs(keys))
	}

	return false, nil
}

func (dao *Dao) RemovePage(ctx context.Context, id string) error {

	var err error

	children, _, err := dao.SelectChildrenPage(ctx, id, NoLimitCursor, 0, false)
	if err != nil {
		return fmt.Errorf("Datastore Error SelectChildPages child page[%v]", err)
	}

	if children != nil {
		return fmt.Errorf("Exist child page[%s]", id)
	}

	cli, err := dao.createClient(ctx)
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

		if dao.ExistFile(ctx, id) {
			return dao.RemoveFile(ctx, id)
		}

		//TODO すべてを削除
		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}
	return nil
}

func (dao *Dao) PutPageSequence(ctx context.Context, ids string, enables string, verCsv string) error {

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

	cli, err := dao.createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	pages := make([]*Page, len(keys))

	err = cli.GetMulti(ctx, keys, pages)
	if err != nil {
		return err
	}

	for idx, page := range pages {
		page.SetTargetVersion(versions[idx])
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

func (dao *Dao) SelectReferencePages(ctx context.Context, id string, typ int) ([]Page, error) {

	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	var pages []Page
	filter := "SiteTemplate="
	if typ == 2 {
		filter = "PageTemplate="
	}

	q := datastore.NewQuery(KindPageName).Filter(filter, id)
	_, err = cli.GetAll(ctx, q, &pages)
	if err != nil {
		return nil, xerrors.Errorf("GetAll() error: %w", err)
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
	return datastore.NameKey(KindPageDataName, id, createSiteKey())
}

func (dao *Dao) SelectPageData(ctx context.Context, id string) (*PageData, error) {

	page := PageData{}
	key := CreatePageDataKey(id)

	cli, err := dao.createClient(ctx)
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

func (dao *Dao) GetPageData(ctx context.Context, ids ...string) ([]PageData, error) {

	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	keys := make([]*datastore.Key, 0, len(ids))
	for _, id := range ids {
		key := CreatePageDataKey(id)
		keys = append(keys, key)
	}

	data := make([]PageData, len(keys))
	err = cli.GetMulti(ctx, keys, data)
	if err != nil {
		return nil, xerrors.Errorf("client.GetAll() error: %w", err)
	}
	return data, nil
}

type Tree struct {
	Page     *Page
	Children []*Tree
	Indent   int
}

func (dao *Dao) CreatePagesTree(ctx context.Context) (*Tree, error) {

	cli, err := dao.createClient(ctx)
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
		if key.Name == ErrorPageID {
			continue
		}
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
