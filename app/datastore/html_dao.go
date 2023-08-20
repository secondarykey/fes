package datastore

import (
	"context"
	"errors"

	"fmt"
	"time"

	"golang.org/x/xerrors"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"cloud.google.com/go/datastore"
)

func (dao *Dao) GetHTML(ctx context.Context, id string) (*HTML, error) {

	var err error
	key := GetHTMLKey(id)
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

func (dao *Dao) PutHTML(ctx context.Context, htmls []*HTML) error {

	cli, err := dao.createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	dsts := make([][]HasKey, 0)
	data := make([]HasKey, 0)

	//TODO ループじゃなくてよい

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

	//HTML数取得
	for _, dst := range dsts {
		_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
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
		key := GetHTMLKey(id)
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

func (dao *Dao) GetHTMLs(ctx context.Context) ([]string, error) {

	cli, err := dao.createClient(ctx, option.WithGRPCDialOption(grpc.WithMaxMsgSize(1024*1024*1000)))
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	var htmls []HTML

	q := datastore.NewQuery(KindHTMLName)
	_, err = cli.GetAll(ctx, q, &htmls)
	if err != nil {
		return nil, xerrors.Errorf("GetAll() error: %w", err)
	}

	ids := make([]string, len(htmls))
	for idx, html := range htmls {
		ids[idx] = html.Key.Name
	}
	return ids, nil
}
