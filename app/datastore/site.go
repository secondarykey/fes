package datastore

import (
	"context"
	"errors"
	"fmt"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/xerrors"

	"cloud.google.com/go/datastore"
)

const KindSiteName = "Site"
const SiteEntityKey = "fixing"

var (
	SiteNotFoundError = fmt.Errorf("site not found")
)

type Site struct {
	Name        string
	Description string
	Root        string
	Managers    []string

	TargetVersion string `datastore:"-"`
	Meta

	//Deprecated
	HTMLCache     bool
	TemplateCache bool
	FileCache     bool
	PageCache     bool
}

func (s *Site) Load(props []datastore.Property) error {
	return datastore.LoadStruct(s, props)
}

func (s *Site) Save() ([]datastore.Property, error) {
	s.update(s.TargetVersion)
	return datastore.SaveStruct(s)
}

func createSiteKey() *datastore.Key {
	return datastore.NameKey(KindSiteName, SiteEntityKey, nil)
}

func (dao *Dao) PutSite(ctx context.Context, site *Site) error {

	var page *Page
	if site.Version == 0 {
		page = &Page{
			Name:   "最初のページ",
			Parent: "",
		}
		page.Deleted = true
		uid := uuid.NewV4()
		page.LoadKey(CreatePageKey(uid.String()))
	}

	cli, err := dao.createClient(ctx)
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

		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}

	return nil
}

var cacheSite *Site

func (dao *Dao) SelectSite(ctx context.Context, version int) (*Site, error) {

	//バージョン指定がない場合
	if version < 0 {
		if cacheSite != nil {
			return cacheSite, nil
		}
	}

	cli, err := dao.createClient(ctx)
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
	return &site, nil
}
