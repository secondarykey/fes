package datastore

import (
	"context"
	"errors"

	"cloud.google.com/go/datastore"
	"golang.org/x/xerrors"
	"google.golang.org/api/iterator"
)

const KindVariableName = "Variable"

type Variable struct {
	TargetVersion string `datastore:"-"`
	Meta
}

func (t *Variable) Load(props []datastore.Property) error {
	return datastore.LoadStruct(t, props)
}

func (t *Variable) Save() ([]datastore.Property, error) {
	t.update(t.TargetVersion)
	return datastore.SaveStruct(t)
}

func createVariableKey(id string) *datastore.Key {
	return datastore.NameKey(KindVariableName, id, createSiteKey())
}

func SelectVariables(ctx context.Context, cur string) ([]Variable, string, error) {

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

	cli, err := createClient(ctx)
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

func SelectVariable(ctx context.Context, id string) (*Variable, error) {

	key := createVariableKey(id)
	cli, err := createClient(ctx)
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

const KindVariableDataName = "VariableData"

type VariableData struct {
	Key     *datastore.Key `datastore:"__key__"`
	Content []byte         `datastore:",noindex"`
}

func (d *VariableData) GetKey() *datastore.Key {
	return d.Key
}

func (d *VariableData) LoadKey(k *datastore.Key) error {
	d.Key = k
	return nil
}

func createVariableDataKey(id string) *datastore.Key {
	return datastore.NameKey(KindVariableDataName, id, createSiteKey())
}

func SelectVariableData(ctx context.Context, id string) (*VariableData, error) {

	key := createVariableDataKey(id)
	cli, err := createClient(ctx)
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

func RemoveVariable(ctx context.Context, id string) error {

	var err error

	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		key := createVariableKey(id)
		err = tx.Delete(key)
		if err != nil {
			return xerrors.Errorf("Variable Delete() error: %w", err)
		}

		dataKey := createVariableDataKey(id)
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

func PutVariable(ctx context.Context, id string, value string, version string) error {

	var err error

	variKey := createVariableKey(id)
	variDataKey := createVariableDataKey(id)

	vari := Variable{}
	variData := VariableData{}

	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	//TODO Version

	vari.LoadKey(variKey)
	variData.LoadKey(variDataKey)

	variData.Content = []byte(value)

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		_, err = tx.Put(vari.GetKey(), &vari)
		if err != nil {
			return xerrors.Errorf("Variable Put() error: %w", err)
		}

		_, err = tx.Put(variData.GetKey(), &variData)
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

func GetVariable(ctx context.Context, key string) (string, error) {

	data, err := SelectVariableData(ctx, key)
	if err != nil {
		return "", xerrors.Errorf("SelectVariableData() error: %w", err)
	}
	return string(data.Content), nil
}
