package datastore

import (
	"context"
	"errors"

	"cloud.google.com/go/datastore"
	"golang.org/x/xerrors"
	"google.golang.org/api/iterator"
)

type VariableSet struct {
	ID           string
	Variable     *Variable
	VariableData *VariableData
}

func (dao *Dao) SelectVariables(ctx context.Context, cur string) ([]Variable, string, error) {

	var rtn []Variable

	q := datastore.NewQuery(KindVariableName).Order("- UpdatedAt")

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
		var vari Variable
		key, err := t.Next(&vari)
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, "", xerrors.Errorf("Template Next() error: %w", err)
		}

		vari.LoadKey(key)
		rtn = append(rtn, vari)
	}

	cursor, err := t.Cursor()
	if err != nil {
		return nil, "", xerrors.Errorf("Template Cursor() error: %w", err)
	}
	return rtn, cursor.String(), nil
}

func (dao *Dao) SelectVariable(ctx context.Context, id string) (*Variable, error) {

	key := getVariableKey(id)
	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	var vari Variable
	err = cli.Get(ctx, key, &vari)
	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
			return nil, xerrors.Errorf("Variable Get() error: %w", err)
		} else {
			return nil, nil
		}
	}

	return &vari, nil
}

func (dao *Dao) SelectVariableData(ctx context.Context, id string) (*VariableData, error) {

	key := getVariableDataKey(id)
	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	var vari VariableData
	err = cli.Get(ctx, key, &vari)
	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
			return nil, xerrors.Errorf("VariableData Get() error: %w", err)
		} else {
			return nil, nil
		}
	}

	return &vari, nil
}

func (dao *Dao) RemoveVariable(ctx context.Context, id string) error {

	var err error

	cli, err := dao.createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		key := getVariableKey(id)
		err = tx.Delete(key)
		if err != nil {
			return xerrors.Errorf("Variable Delete() error: %w", err)
		}

		dataKey := getVariableDataKey(id)
		err = tx.Delete(dataKey)
		if err != nil {
			return xerrors.Errorf("VariableData Delete() error: %w", err)
		}
		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}

	return nil
}

func (dao *Dao) PutVariable(ctx context.Context, vs *VariableSet) error {

	var err error

	variKey := getVariableKey(vs.ID)
	variDataKey := getVariableDataKey(vs.ID)

	vari := vs.Variable
	variData := vs.VariableData

	vari.LoadKey(variKey)
	variData.LoadKey(variDataKey)

	cli, err := dao.createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		_, err = tx.Put(vari.GetKey(), vari)
		if err != nil {
			return xerrors.Errorf("Variable Put() error: %w", err)
		}

		_, err = tx.Put(variData.GetKey(), variData)
		if err != nil {
			return xerrors.Errorf("VariableData Put() error: %w", err)
		}
		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}
	return nil
}

func (dao *Dao) GetVariable(ctx context.Context, key string) (string, error) {

	data, err := dao.SelectVariableData(ctx, key)
	if err != nil {
		return "", xerrors.Errorf("SelectVariableData() error: %w", err)
	}
	return string(data.Content), nil
}
