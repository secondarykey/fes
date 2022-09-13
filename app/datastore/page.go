package datastore

import (
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/xerrors"

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

func CreatePageKey() *datastore.Key {
	uid := uuid.NewV4()
	return GetPageKey(uid.String())
}

func GetPageKey(id string) *datastore.Key {
	return datastore.NameKey(KindPageName, id, getSiteKey())
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

func GetPageDataKey(id string) *datastore.Key {
	return datastore.NameKey(KindPageDataName, id, getSiteKey())
}
