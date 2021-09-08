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

const (
	FileTypeData      = 1
	FileTypePageImage = 2
	FileTypeSystem    = 3
)

const (
	SystemFaviconID        = "system-favicon"
	draftPageImageIDSuffix = "DRAFT"
)

const KindFileName = "File"

type File struct {
	Size int64
	Type int

	TargetVersion string `datastore:"-"`
	Meta
}

func (f *File) Load(props []datastore.Property) error {
	return datastore.LoadStruct(f, props)
}

func (f *File) Save() ([]datastore.Property, error) {
	f.update(f.TargetVersion)
	return datastore.SaveStruct(f)
}

func createFileKey(name string) *datastore.Key {
	return datastore.NameKey(KindFileName, name, createSiteKey())
}

func GetAllFiles(ctx context.Context) ([]*File, error) {

	var dst []*File
	q := datastore.NewQuery(KindFileName)

	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.GetAll(ctx, q, &dst)
	if err != nil {
		return nil, xerrors.Errorf("File GetAll() error: %w", err)
	}
	return dst, nil
}

func SelectFiles(ctx context.Context, tBuf string, cur string) ([]File, string, error) {

	var s []File

	typ := 0
	if tBuf == "1" || tBuf == "2" {
		typ, _ = strconv.Atoi(tBuf)
	}

	cli, err := createClient(ctx)
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

func SelectFile(ctx context.Context, name string) (*File, error) {

	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	rtn := File{}
	key := createFileKey(name)

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

func SaveFile(ctx context.Context, fs *FileSet) error {

	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	id := fs.ID
	if id == "" {
		id = fs.Name
	}

	f := fs.File
	fd := fs.FileData

	f.LoadKey(createFileKey(id))
	fd.LoadKey(createFileDataKey(id))

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

func PutFileData(ctx context.Context, id string, data []byte, mime string) error {

	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		fileKey := createFileKey(id)

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
		fileData.LoadKey(createFileDataKey(id))
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

func ExistFile(ctx context.Context, id string) bool {

	file := &File{}
	file.Key = createFileKey(id)

	cli, err := createClient(ctx)
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

func RemoveFile(ctx context.Context, id string) error {

	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		fkey := createFileKey(id)
		err := tx.Delete(fkey)
		if err != nil {
			return err
		}
		fdkey := createFileDataKey(id)
		return tx.Delete(fdkey)
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}

	return nil
}

const KindFileDataName = "FileData"

type FileData struct {
	Key     *datastore.Key `datastore:"__key__"`
	Mime    string
	Content []byte `datastore:",noindex"`
}

func (d *FileData) GetKey() *datastore.Key {
	return d.Key
}

func (d *FileData) LoadKey(k *datastore.Key) error {
	d.Key = k
	return nil
}

func createFileDataKey(name string) *datastore.Key {
	return datastore.NameKey(KindFileDataName, name, createSiteKey())
}

func GetFileData(ctx context.Context, name string) (*FileData, error) {

	var rtn FileData
	key := createFileDataKey(name)

	cli, err := createClient(ctx)
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

func GetFavicon(ctx context.Context) ([]byte, error) {
	d, err := GetFileData(ctx, SystemFaviconID)
	if err != nil {
		return nil, xerrors.Errorf("GetFileData() error: %w", err)
	}
	if d == nil {
		return nil, nil
	}
	return d.Content, nil
}

type FileSet struct {
	ID       string
	Name     string
	File     *File
	FileData *FileData
}

func GetFileSet(tx *datastore.Transaction, id string) (*FileSet, error) {

	fkey := createFileKey(id)
	fdkey := createFileDataKey(id)

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

	fmt.Println("PublishPageImage()")

	draftId := CreateDraftPageImageID(id)

	fmt.Println("DraftID", draftId)

	fs, err := GetFileSet(tx, draftId)
	if err != nil {
		return xerrors.Errorf("GetFileSet() error: %w", err)
	}

	if fs == nil {
		fmt.Println("fs is nil")
		return nil
	}

	f := fs.File
	fd := fs.FileData

	ids := make([]*datastore.Key, 2)
	ids[0] = f.GetKey()
	ids[1] = fd.GetKey()

	f.LoadKey(createFileKey(id))
	fd.LoadKey(createFileDataKey(id))

	fmt.Println("delete multi", ids)
	err = tx.DeleteMulti(ids)
	if err != nil {
		return xerrors.Errorf("DeleteMulti() error: %w", err)
	}

	fmt.Println("file put", f.GetKey())
	_, err = tx.Put(f.GetKey(), f)
	if err != nil {
		return xerrors.Errorf("File Put() error: %w", err)
	}

	fmt.Println("filedata put", fd.GetKey())
	_, err = tx.Put(fd.GetKey(), fd)
	if err != nil {
		return xerrors.Errorf("FileData Put() error: %w", err)
	}

	return nil
}
