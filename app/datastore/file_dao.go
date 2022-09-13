package datastore

import (
	"context"
	"fmt"

	"errors"
	_ "image/gif"
	_ "image/png"
	"strconv"

	"golang.org/x/xerrors"
	"google.golang.org/api/iterator"

	"cloud.google.com/go/datastore"
)

type FileSet struct {
	ID       string
	Name     string
	File     *File
	FileData *FileData
}

func (dao *Dao) GetAllFiles(ctx context.Context) ([]*File, error) {

	var dst []*File
	q := datastore.NewQuery(KindFileName)

	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.GetAll(ctx, q, &dst)
	if err != nil {
		return nil, xerrors.Errorf("File GetAll() error: %w", err)
	}
	return dst, nil
}

func (dao *Dao) SelectFiles(ctx context.Context, tBuf string, cur string) ([]File, string, error) {

	var s []File

	typ := 0
	if tBuf == "1" || tBuf == "2" {
		typ, _ = strconv.Atoi(tBuf)
	}

	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, "", xerrors.Errorf("createClient() error: %w", err)
	}

	q := datastore.NewQuery(KindFileName).Order("- UpdatedAt")

	if typ == FileTypeData || typ == FileTypePageImage {
		q = q.Filter("Type=", typ)
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

	t := cli.Run(ctx, q)
	for {
		var f File
		key, err := t.Next(&f)

		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, "", xerrors.Errorf("File Next() error: %w", err)
		}
		f.LoadKey(key)
		s = append(s, f)
	}

	cursor, err := t.Cursor()
	if err != nil {
		return nil, "", xerrors.Errorf("File Cursor() error: %w", err)
	}

	return s, cursor.String(), nil
}

func (dao *Dao) SelectFile(ctx context.Context, name string) (*File, error) {

	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	rtn := File{}
	key := getFileKey(name)

	err = cli.Get(ctx, key, &rtn)
	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
			return nil, xerrors.Errorf("File Get() error: %w", err)
		} else if errors.Is(err, datastore.ErrNoSuchEntity) {
			return nil, nil
		}
	}
	return &rtn, nil
}

func (dao *Dao) SaveFile(ctx context.Context, fs *FileSet) error {

	cli, err := dao.createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	id := fs.ID
	if id == "" {
		id = fs.Name
	}

	f := fs.File
	fd := fs.FileData

	f.LoadKey(getFileKey(id))
	fd.LoadKey(getFileDataKey(id))

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		_, err = tx.Put(f.GetKey(), f)
		if err != nil {
			return xerrors.Errorf("File Put() error: %w", err)
		}

		_, err = tx.Put(fd.GetKey(), fd)
		if err != nil {
			return xerrors.Errorf("FileData Put() error: %w", err)
		}
		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}

	return nil
}

func (dao *Dao) PutFileData(ctx context.Context, id string, data []byte, mime string) error {

	cli, err := dao.createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		fileKey := getFileKey(id)

		file := File{}
		err := tx.Get(fileKey, &file)
		if err != nil {
			if !errors.Is(err, datastore.ErrNoSuchEntity) {
				return err
			}
		}

		file.Size = int64(len(data))
		_, err = tx.Put(fileKey, &file)
		if err != nil {
			return err
		}

		fileData := &FileData{
			Content: data,
			Mime:    mime,
		}
		fileData.LoadKey(getFileDataKey(id))
		_, err = tx.Put(fileData.GetKey(), fileData)
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

func (dao *Dao) ExistFile(ctx context.Context, id string) bool {

	file := &File{}
	file.Key = getFileKey(id)

	cli, err := dao.createClient(ctx)
	if err != nil {
		return false
	}

	err = cli.Get(ctx, file.Key, file)
	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
			return true
		} else if errors.Is(err, datastore.ErrNoSuchEntity) {
			return false
		}

		//TODO log
	}
	return true
}

func (dao *Dao) RemoveFile(ctx context.Context, id string) error {

	cli, err := dao.createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		fkey := getFileKey(id)
		err := tx.Delete(fkey)
		if err != nil {
			return err
		}
		fdkey := getFileDataKey(id)
		return tx.Delete(fdkey)
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}

	return nil
}

func (dao *Dao) GetFileData(ctx context.Context, name string) (*FileData, error) {

	var rtn FileData
	key := getFileDataKey(name)

	cli, err := dao.createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient error: %w", err)
	}

	err = cli.Get(ctx, key, &rtn)
	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
			return nil, xerrors.Errorf("FileData Get() error: %w", err)
		} else if errors.Is(err, datastore.ErrNoSuchEntity) {
			return nil, nil
		}
	}
	return &rtn, nil
}

func (dao *Dao) GetFavicon(ctx context.Context) ([]byte, error) {
	d, err := dao.GetFileData(ctx, SystemFaviconID)
	if err != nil {
		return nil, xerrors.Errorf("GetFileData() error: %w", err)
	}
	if d == nil {
		return nil, nil
	}
	return d.Content, nil
}

func GetFileSet(tx *datastore.Transaction, id string) (*FileSet, error) {

	fkey := getFileKey(id)
	fdkey := getFileDataKey(id)

	var f File
	var fd FileData

	err := tx.Get(fkey, &f)
	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
			return nil, xerrors.Errorf("File Get() error: %w", err)
		}
		return nil, nil
	}

	err = tx.Get(fdkey, &fd)
	if err != nil {
		return nil, xerrors.Errorf("FileData Get() error: %w", err)
	}

	var fs FileSet
	fs.ID = id
	fs.File = &f
	fs.FileData = &fd
	return &fs, nil
}

func CreateDraftPageImageID(id string) string {
	return fmt.Sprintf("%s-%s", id, draftPageImageIDSuffix)
}

func PublishPageImage(tx *datastore.Transaction, id string) error {

	draftId := CreateDraftPageImageID(id)

	fs, err := GetFileSet(tx, draftId)
	if err != nil {
		return xerrors.Errorf("GetFileSet() error: %w", err)
	}

	if fs == nil {
		return nil
	}

	f := fs.File
	fd := fs.FileData

	ids := make([]*datastore.Key, 2)
	ids[0] = f.GetKey()
	ids[1] = fd.GetKey()

	f.LoadKey(getFileKey(id))
	fd.LoadKey(getFileDataKey(id))

	err = tx.DeleteMulti(ids)
	if err != nil {
		return xerrors.Errorf("DeleteMulti() error: %w", err)
	}

	_, err = tx.Put(f.GetKey(), f)
	if err != nil {
		return xerrors.Errorf("File Put() error: %w", err)
	}

	_, err = tx.Put(fd.GetKey(), fd)
	if err != nil {
		return xerrors.Errorf("FileData Put() error: %w", err)
	}

	return nil
}
