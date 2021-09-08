package datastore

import (
	"context"
	"errors"

	"fmt"
	"time"

	"golang.org/x/xerrors"

	"cloud.google.com/go/datastore"
)

const KindHTMLName = "HTML"

type HTML struct {
	Content       []byte `datastore:",noindex"`
	Children      int    //ignore
	PageKey       string //added
	TargetVersion string `datastore:"-"`
	Meta
}

func (h *HTML) Load(props []datastore.Property) error {
	return datastore.LoadStruct(h, props)
}

func (h *HTML) Save() ([]datastore.Property, error) {
	h.update(h.TargetVersion)
	return datastore.SaveStruct(h)
}

func CreateHTMLKey(id string) *datastore.Key {
	return datastore.NameKey(KindHTMLName, id, createSiteKey())
}

func GetHTML(ctx context.Context, id string) (*HTML, error) {

	var err error
	key := CreateHTMLKey(id)
	html := HTML{}

	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	err = cli.Get(ctx, key, &html)
	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
			return nil, err
		} else {
			return nil, nil
		}
	}
	return &html, nil
}

func PutHTML(ctx context.Context, htmls []*HTML, page *Page) error {

	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	dsts := make([]HasKey, len(htmls))
	for idx, html := range htmls {
		dsts[idx] = html
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		err = PublishPageImage(tx, page.GetKey().Name)
		if err != nil {
			return xerrors.Errorf("PublishPageImage() error: %w", err)
		}

		err = PutMulti(tx, dsts)
		if err != nil {
			return xerrors.Errorf("HTML PutMulti() error: %w", err)
		}

		if page != nil {
			if page.Publish.IsZero() {
				page.Publish = time.Now()
			}
			err = Put(tx, page)
			if err != nil {
				return xerrors.Errorf("Page(Publish) Put() error: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		return xerrors.Errorf("PutHTML() transaction error: %w", err)
	}

	return nil
}

func RemoveHTML(ctx context.Context, id string) error {

	page, err := SelectPage(ctx, id, -1)
	if err != nil {
		return err
	}
	if page == nil {
		return fmt.Errorf("page not found[%s]", id)
	}

	//TODO ページ数個削除
	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		key := CreateHTMLKey(id)
		err = tx.Delete(key)
		if err != nil {
			return err
		}

		page.Publish = time.Time{}
		_, err = tx.Put(page.GetKey(), page)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}

	return nil
}
