package datastore

import (
	"app/config"
	"context"
	"errors"
	"sync"

	"cloud.google.com/go/datastore"
	"golang.org/x/xerrors"
	"google.golang.org/api/option"
)

const NoLimitCursor = "NoLimit"

type Dao struct {
}

func NewDao() *Dao {
	var dao Dao
	return &dao
}

// クライアントはスレッドセーフなためアプリ全体で1つを共有する。
// リクエスト毎に生成すると gRPC コネクション確立のコストが毎回かかる。
var (
	cliOnce   sync.Once
	sharedCli *datastore.Client
	cliErr    error
)

// Close は互換性のために残している。クライアントは共有のため何もしない。
func (dao *Dao) Close() error {
	return nil
}

//
// GRPC Large
// cli, err := createClient(ctx, option.WithGRPCDialOption(grpc.WithMaxMsgSize(10_000_000)))
//
func (dao *Dao) createClient(ctx context.Context, opts ...option.ClientOption) (*datastore.Client, error) {
	cliOnce.Do(func() {
		c := config.Get()
		// リクエストの context と寿命を切り離すため Background を使用する
		sharedCli, cliErr = datastore.NewClient(context.Background(), c.ProjectID, opts...)
	})
	if cliErr != nil {
		return nil, xerrors.Errorf("datastore.CreateClient() error: %w", cliErr)
	}
	return sharedCli, nil
}

// IsNoSuchEntity は err が「エンティティが存在しない」エラーのみで構成されているかを返す。
// GetMulti が返す MultiError で一部のみ欠損している場合も true になる。
// この場合、結果スライスの欠損分はゼロ値(Key が nil)のままになっている。
func IsNoSuchEntity(err error) bool {
	if err == nil {
		return false
	}
	var merr datastore.MultiError
	if errors.As(err, &merr) {
		for _, e := range merr {
			if e != nil && !errors.Is(e, datastore.ErrNoSuchEntity) {
				return false
			}
		}
		return true
	}
	return errors.Is(err, datastore.ErrNoSuchEntity)
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
