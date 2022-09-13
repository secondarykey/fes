package datastore

import (
	"context"
	"errors"
	"log"
	"strconv"

	"cloud.google.com/go/datastore"
	"golang.org/x/xerrors"
	"google.golang.org/api/iterator"
)

type TemplateSet struct {
	ID           string
	Template     *Template
	TemplateData *TemplateData
}

func (dao *Dao) PutTemplate(ctx context.Context, ts *TemplateSet) error {

	var err error

	cli, err := dao.createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		_, err = tx.Put(ts.Template.GetKey(), ts.Template)
		if err != nil {
			return xerrors.Errorf("Template Put() error: %w", err)
		}

		_, err = tx.Put(ts.TemplateData.GetKey(), ts.TemplateData)
		if err != nil {
			return xerrors.Errorf("TemplateData Put() error: %w", err)
		}
		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}
	return nil
}

func (dao *Dao) SelectTemplate(ctx context.Context, id string) (*Template, error) {
	temp := Template{}

	//Method
	key := GetTemplateKey(id)
	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	err = cli.Get(ctx, key, &temp)
	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
			return nil, xerrors.Errorf("Template Get() error: %w", err)
		} else {
			return nil, nil
		}
	}
	return &temp, nil
}

func (dao *Dao) SelectTemplates(ctx context.Context, ty string, cur string) ([]Template, string, error) {

	var rtn []Template

	q := datastore.NewQuery(KindTemplateName).Order("- UpdatedAt")

	if ty != "all" {
		v, err := strconv.Atoi(ty)
		if err == nil {
			q = q.Filter("Type=", v)
		} else {
			log.Println("strconv parse error", ty)
		}
	}

	if cur != NoLimitCursor {
		q = q.Limit(10)
		if cur != "" {
			cursor, err := datastore.DecodeCursor(cur)
			if err != nil {
				return nil, "", xerrors.Errorf("datastore.DecodeCursor() error: %w", err)
			}
			q = q.Start(cursor)
		}
	}

	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, "", xerrors.Errorf("createClient() error: %w", err)
	}

	t := cli.Run(ctx, q)
	for {
		var tmp Template
		key, err := t.Next(&tmp)
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, "", xerrors.Errorf("Template Next() error: %w", err)
		}
		tmp.LoadKey(key)
		rtn = append(rtn, tmp)
	}

	cursor, err := t.Cursor()
	if err != nil {
		return nil, "", xerrors.Errorf("Template Cursor() error: %w", err)
	}

	return rtn, cursor.String(), nil
}

func (dao *Dao) SelectTemplateData(ctx context.Context, id string) (*TemplateData, error) {
	temp := TemplateData{}

	//Method
	key := GetTemplateDataKey(id)
	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	err = cli.Get(ctx, key, &temp)
	if err != nil {
		return nil, xerrors.Errorf("TemplateData Get() error: %w", err)
	}
	return &temp, nil
}

func (dao *Dao) RemoveTemplate(ctx context.Context, id string) error {

	var err error

	cli, err := dao.createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		key := GetTemplateKey(id)
		err = tx.Delete(key)
		if err != nil {
			return xerrors.Errorf("Template Delete() error: %w", err)
		}

		dataKey := GetTemplateDataKey(id)
		err = tx.Delete(dataKey)
		if err != nil {
			return xerrors.Errorf("TemplateData Delete() error: %w", err)
		}
		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}

	return nil
}
