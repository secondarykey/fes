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
	Content  []byte `datastore:",noindex"`
	Children int    //ignore
	PageKey  string //added
	Meta
}

func (h *HTML) Load(props []datastore.Property) error {
	return datastore.LoadStruct(h, props)
}

func (h *HTML) Save() ([]datastore.Property, error) {
	err := h.update()
	if err != nil {
		return nil, xerrors.Errorf("Meta update() error: %w", err)
	}
	return datastore.SaveStruct(h)
}

func CreateHTMLKey(id string) *datastore.Key {
	return datastore.NameKey(KindHTMLName, id, createSiteKey())
}

func (dao *Dao) GetHTML(ctx context.Context, id string) (*HTML, error) {

	var err error
	key := CreateHTMLKey(id)
	html := HTML{}

	cli, err := dao.createClient(ctx)
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

const MB4 = 3_900_000

func (dao *Dao) PutHTML(ctx context.Context, htmls []*HTML, page *Page) error {

	cli, err := dao.createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	dsts := make([][]HasKey, 0)
	data := make([]HasKey, 0)

	sz := 0
	for idx, html := range htmls {
		data = append(data, html)
		sz += len(html.Content)
		if sz > MB4 || idx+1 == len(htmls) {
			dsts = append(dsts, data)
			data = make([]HasKey, 0)
			sz = 0
		}
	}

	for _, dst := range dsts {
		_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

			if page != nil {
				err = PublishPageImage(tx, page.GetKey().Name)
				if err != nil {
					return xerrors.Errorf("PublishPageImage() error: %w", err)
				}
				if page.Publish.IsZero() {
					page.Publish = time.Now()
					err = Put(tx, page)
					if err != nil {
						return xerrors.Errorf("Page(Publish) Put() error: %w", err)
					}
				}
			}

			err = PutMulti(tx, dst)
			if err != nil {
				return xerrors.Errorf("HTML PutMulti() error: %w", err)
			}

			return nil
		})
		if err != nil {
			return xerrors.Errorf("PutHTML() transaction error: %w", err)
		}
	}

	return nil
}

func (dao *Dao) RemoveHTML(ctx context.Context, id string) error {

	page, err := dao.SelectPage(ctx, id, -1)
	if err != nil {
		return err
	}
	if page == nil {
		return fmt.Errorf("page not found[%s]", id)
	}

	//TODO ページ数個削除
	cli, err := dao.createClient(ctx)
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
