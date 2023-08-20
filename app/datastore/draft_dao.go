package datastore

import (
	"context"
	"errors"

	"cloud.google.com/go/datastore"
	"golang.org/x/xerrors"
	"google.golang.org/api/iterator"
)

type DraftSet struct {
	Draft *Draft
	Pages []*DraftPage
}

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

func (dao *Dao) PutDraftSet(ctx context.Context, set *DraftSet) error {

	var err error

	cli, err := dao.createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	existSet, err := dao.SelectDraftSet(ctx, set.Draft.Key.Name)
	if err != nil {
		return xerrors.Errorf("selectDraftSet() error: %w", err)
	}

	pages, err := copyDraft(existSet.Pages, set.Pages)
	if err != nil {
		return xerrors.Errorf("copyDraft() error: %w", err)
	}

	has := make([]HasKey, len(pages))
	for idx, elm := range pages {
		has[idx] = elm
	}
	keys := getKeys(has)

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		draft := set.Draft

		_, err = tx.Put(draft.GetKey(), draft)
		if err != nil {
			return xerrors.Errorf("Draft Put() error: %w", err)
		}

		_, err = tx.PutMulti(keys, pages)
		if err != nil {
			return xerrors.Errorf("Draft Put() error: %w", err)
		}

		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}
	return nil
}

func copyDraft(exists []*DraftPage, forms []*DraftPage) ([]*DraftPage, error) {

	if len(exists) != len(forms) {
		return nil, xerrors.Errorf("DraftPages length error[%d != %d]", len(exists), len(forms))
	}

	//名寄せ
	for _, p := range forms {
		exist := false
		for _, e := range exists {
			if p.Key.Name == e.Key.Name {
				exist = true
				e.Seq = p.Seq
				e.PublishUpdate = p.PublishUpdate
				break
			}
		}

		if !exist {
			return nil, xerrors.Errorf("DraftPages NotFound error[%s]", p.Key.Name)
		}
	}
	return exists, nil
}

func (dao *Dao) SelectDraft(ctx context.Context, id string) (*Draft, error) {

	var err error
	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	var draft Draft

	key := GetDraftKey(id)
	err = cli.Get(ctx, key, &draft)
	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
			return nil, xerrors.Errorf("Draft Get() error: %w", err)
		} else {
			return nil, nil
		}
	}

	return &draft, nil
}

func (dao *Dao) SelectDraftSet(ctx context.Context, id string) (*DraftSet, error) {

	draft, err := dao.SelectDraft(ctx, id)
	if err != nil {
		return nil, xerrors.Errorf("SelectDraft() error: %w", err)
	}

	pages, err := dao.SelectDraftPages(ctx, id)
	if err != nil {
		return nil, xerrors.Errorf("SelectDraftPages() error: %w", err)
	}

	set := DraftSet{}

	set.Draft = draft
	set.Pages = pages

	return &set, nil
}

func (dao *Dao) getDraftPageKeys(ctx context.Context, id string) ([]*datastore.Key, error) {

	var err error
	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	q := datastore.NewQuery(KindDraftPageName).Filter("DraftID=", id).KeysOnly()
	keys, err := cli.GetAll(ctx, q, nil)
	if err != nil {
		return nil, xerrors.Errorf("GetAll() error: %w", err)
	}
	return keys, nil
}

func (dao *Dao) SelectDraftPages(ctx context.Context, id string) ([]*DraftPage, error) {

	var err error
	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	q := datastore.NewQuery(KindDraftPageName).Filter("DraftID=", id).Order("Seq").Order("- CreatedAt")

	t := cli.Run(ctx, q)
	var pages []*DraftPage
	for {
		var page DraftPage
		_, err := t.Next(&page)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, xerrors.Errorf("Draft Next() error: %w", err)
		}
		pages = append(pages, &page)
	}

	return pages, nil
}

func (dao *Dao) SelectDraftPage(ctx context.Context, id string) (*DraftPage, error) {

	var err error
	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	var page DraftPage

	key := GetDraftPageKey(id)
	err = cli.Get(ctx, key, &page)
	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
			return nil, xerrors.Errorf("DraftPage Get() error: %w", err)
		} else {
			return nil, nil
		}
	}
	return &page, nil
}

func (dao *Dao) RemoveDraft(ctx context.Context, id string) error {
	var err error
	var keys []*datastore.Key

	cli, err := dao.createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	keys, err = dao.getDraftPageKeys(ctx, id)
	if err != nil {
		return xerrors.Errorf("getDraftPageKeys() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		key := GetDraftKey(id)
		err = tx.Delete(key)
		if err != nil {
			return xerrors.Errorf("Draft Delete() error: %w", err)
		}

		if len(keys) > 0 {
			err = tx.DeleteMulti(keys)
			if err != nil {
				return xerrors.Errorf("DraftPage Delete() error: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}

	return nil

}

func (dao *Dao) AddDraftPage(ctx context.Context, draftId string, pageId string) error {

	cli, err := dao.createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	//ページ検索
	p, err := dao.SelectPage(ctx, pageId, -1)
	if err != nil {
		return xerrors.Errorf("SelectPage() error: %w", err)
	}

	//全件検索
	pages, err := dao.SelectDraftPages(ctx, draftId)
	if err != nil {
		return xerrors.Errorf("SelectDraftPages() error: %w", err)
	}

	//件数の番号で追加
	seq := len(pages) + 1
	var page DraftPage
	page.LoadKey(CreateDraftPageKey())

	page.DraftID = draftId
	page.PageID = pageId
	page.Name = p.Name
	page.Seq = seq

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		_, err = tx.Put(page.GetKey(), &page)
		if err != nil {
			return xerrors.Errorf("DraftPage Put() error: %w", err)
		}
		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}

	return nil
}

func (dao *Dao) RemoveDraftPage(ctx context.Context, id string) (string, error) {

	cli, err := dao.createClient(ctx)
	if err != nil {
		return "", xerrors.Errorf("createClient() error: %w", err)
	}

	p, err := dao.SelectDraftPage(ctx, id)
	if err != nil {
		return "", xerrors.Errorf("SelectDraftPages() error: %w", err)
	}

	draftId := p.DraftID
	//全件検索
	pages, err := dao.SelectDraftPages(ctx, draftId)
	if err != nil {
		return "", xerrors.Errorf("SelectDraftPages() error: %w", err)
	}

	//指定のページのキー値を取得
	deleteKey := GetDraftPageKey(id)

	var keys []*datastore.Key
	var data []*DraftPage

	seq := 0
	//番号を設定
	for _, p := range pages {
		if p.Key.Name != id {
			seq++
			if p.Seq != seq {
				p.Seq = seq
				keys = append(keys, p.GetKey())
				data = append(data, p)
			}
		}
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		err = tx.Delete(deleteKey)
		if err != nil {
			return xerrors.Errorf("DraftPage Delete() error: %w", err)
		}
		_, err = tx.PutMulti(keys, data)
		if err != nil {
			return xerrors.Errorf("DraftPage PutMulti() error: %w", err)
		}

		return nil
	})

	if err != nil {
		return "", xerrors.Errorf("tx error: %w", err)
	}

	return draftId, nil
}
