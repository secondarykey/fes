package datastore

import (
	"app/config"
	"context"

	"cloud.google.com/go/datastore"
	"golang.org/x/xerrors"
	"google.golang.org/api/option"
)

const NoLimitCursor = "NoLimit"

func createClient(ctx context.Context, opts ...option.ClientOption) (*datastore.Client, error) {
	c := config.Get()
	cli, err := datastore.NewClient(ctx, c.ProjectID, opts...)

	if err != nil {
		return nil, xerrors.Errorf("datastore.CreateClient() error: %w")
	}
	return cli, nil
}

func PutMulti(tx *datastore.Transaction, dsts []HasKey) error {

	keys := make([]*datastore.Key, len(dsts))
	for idx, elm := range dsts {
		keys[idx] = elm.GetKey()
	}

	_, err := tx.PutMulti(keys, dsts)
	if err != nil {
		return xerrors.Errorf("PutMulti() error: %w", err)
	}

	return nil
}

func Put(tx *datastore.Transaction, dst HasKey) error {
	_, err := tx.Put(dst.GetKey(), dst)
	if err != nil {
		return xerrors.Errorf("Put() error: %w", err)
	}
	return nil
}
