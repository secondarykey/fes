package datastore

import (
	"app/config"
	"context"

	"cloud.google.com/go/datastore"
	"golang.org/x/xerrors"
	"google.golang.org/api/option"
)

const NoLimitCursor = "NoLimit"

type Dao struct {
	cli *datastore.Client
}

func NewDao() *Dao {
	var dao Dao
	return &dao
}

func (dao *Dao) Close() error {
	if dao.cli != nil {
		err := dao.cli.Close()
		if err != nil {
			return xerrors.Errorf("dao Close() error: %w", err)
		}
	}
	return nil
}

//
// GRPC Large
// cli, err := createClient(ctx, option.WithGRPCDialOption(grpc.WithMaxMsgSize(10_000_000)))
//
func (dao *Dao) createClient(ctx context.Context, opts ...option.ClientOption) (*datastore.Client, error) {
	var err error
	if dao.cli == nil {
		c := config.Get()
		dao.cli, err = datastore.NewClient(ctx, c.ProjectID, opts...)
		if err != nil {
			return nil, xerrors.Errorf("datastore.CreateClient() error: %w")
		}
	}
	return dao.cli, nil
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

func getIDs(keys []*datastore.Key) []string {

	ids := make([]string, len(keys))

	for idx, key := range keys {
		ids[idx] = key.Name
	}

	return ids
}

func getKeys(metas []HasKey) []*datastore.Key {
	keys := make([]*datastore.Key, len(metas))
	for idx, meta := range metas {
		keys[idx] = meta.GetKey()
	}
	return keys
}
