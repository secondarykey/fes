package datastore

import (
	"context"
	"errors"

	"cloud.google.com/go/datastore"
	"golang.org/x/xerrors"
	"google.golang.org/api/iterator"
)

func (dao *Dao) SelectDrafts(ctx context.Context, cur string) ([]Draft, string, error) {

	var rtn []Draft

	q := datastore.NewQuery(KindDraftName).Order("- UpdatedAt")

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
		var draft Draft
		key, err := t.Next(&draft)
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, "", xerrors.Errorf("Draft Next() error: %w", err)
		}
		draft.LoadKey(key)
		rtn = append(rtn, draft)
	}

	cursor, err := t.Cursor()
	if err != nil {
		return nil, "", xerrors.Errorf("Template Cursor() error: %w", err)
	}

	return rtn, cursor.String(), nil
}
