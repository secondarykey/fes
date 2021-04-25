package datastore

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
	Name          string
	Description   string
	Root          string
	Managers      []string
	HTMLCache     bool
	TemplateCache bool
	FileCache     bool
	PageCache     bool

	TargetVersion string `datastore:"-"`
	Meta
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

func PutSite(r *http.Request) error {

	ver := r.FormValue("version")
	version, err := strconv.Atoi(ver)
	if err != nil {
		return err
	}

	ctx := r.Context()

	site, foundErr := SelectSite(ctx, version)
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
		page.LoadKey(CreatePageKey(uid.String()))
	}

	cli, err := createClient(ctx)
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
		setDatastoreCache(cacheSite)

		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}

	return nil
}

var cacheSite *Site

func SelectSite(ctx context.Context, version int) (*Site, error) {

	//バージョン指定がない場合
	if version < 0 {
		if cacheSite != nil {
			return cacheSite, nil
		}
	}

	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	key := createSiteKey()

	fmt.Println(key)
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
	setDatastoreCache(cacheSite)
	return &site, nil
}

func setDatastoreCache(site *Site) {
	return
}
