package datastore

import (
	"app/api"

	"bytes"
	"context"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	verr "github.com/knightso/base/errors"
	"github.com/knightso/base/gae/ds"
	"github.com/nfnt/resize"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

const KindFileName = "File"

type File struct {
	Size int64
	Type int
	ds.Meta
}

func createFileKey(r *http.Request, name string) *datastore.Key {
	c := appengine.NewContext(r)
	return datastore.NewKey(c, KindFileName, name, 0, nil)
}

func getFileCursor(p int) string {
	return "file_" + strconv.Itoa(p) + "_cursor"
}

func SelectFiles(r *http.Request, tBuf string, p int) ([]File, error) {

	var s []File

	typ := 0
	if tBuf == "1" || tBuf == "2" {
		typ, _ = strconv.Atoi(tBuf)
	}

	c := appengine.NewContext(r)
	cursor := ""

	//q := datastore.NewQuery(KindFileName).Order("- UpdatedAt")
	q := datastore.NewQuery(KindFileName).Order("- UpdatedAt")

	if typ == api.FileTypeData || typ == api.FileTypePageImage {
		q = q.Filter("Type=", typ)
	}

	if p > 0 {
		item, err := memcache.Get(c, getFileCursor(p))
		if err == nil {
			cursor = string(item.Value)
		}
		q = q.Limit(10)
		if cursor != "" {
			cur, err := datastore.DecodeCursor(cursor)
			if err == nil {
				q = q.Start(cur)
			}
		}
	}

	t := q.Run(c)
	for {
		var f File
		key, err := t.Next(&f)

		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		f.SetKey(key)
		s = append(s, f)
	}

	if p > 0 {
		cur, err := t.Cursor()
		if err != nil {
			return nil, err
		}

		err = memcache.Set(c, &memcache.Item{
			Key:   getFileCursor(p + 1),
			Value: []byte(cur.String()),
		})
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

func SelectFile(r *http.Request, name string) (*File, error) {
	c := appengine.NewContext(r)

	rtn := File{}
	key := createFileKey(r, name)

	err := ds.Get(c, key, &rtn)
	if err != nil {
		if verr.Root(err) != datastore.ErrNoSuchEntity {
			return nil, verr.Root(err)
		} else if verr.Root(err) == datastore.ErrNoSuchEntity {
			return nil, nil
		}
	}
	return &rtn, nil
}

func SaveFile(r *http.Request, id string, t int) error {

	upload, header, err := r.FormFile("file")
	if err != nil {
		return err
	}
	defer upload.Close()

	b, flg, err := convertImage(upload)
	if err != nil {
		return err
	}

	c := appengine.NewContext(r)
	if id == "" {
		id = header.Filename
	}

	option := &datastore.TransactionOptions{XG: true}
	err = datastore.RunInTransaction(c, func(ctx context.Context) error {

		fileKey := createFileKey(r, id)

		file := File{}
		err = ds.Get(c, fileKey, &file)
		if err != nil {
			if verr.Root(err) != datastore.ErrNoSuchEntity {
				return err
			}

			file.Key = fileKey
		}

		file.Size = int64(len(b))
		file.Type = t

		err = ds.Put(ctx, &file)
		if err != nil {
			return err
		}

		mime := header.Header["Content-Type"][0]
		if flg {
			mime = "image/jpeg"
		}

		fileData := &FileData{
			Content: b,
			Mime:    mime,
		}
		fileData.SetKey(createFileDataKey(r, id))
		err = ds.Put(ctx, fileData)
		if err != nil {
			return err
		}
		return nil
	}, option)

	return err
}

func PutFileData(r *http.Request, id string, data []byte, mime string) error {

	option := &datastore.TransactionOptions{XG: true}
	c := appengine.NewContext(r)
	return datastore.RunInTransaction(c, func(ctx context.Context) error {

		fileKey := createFileKey(r, id)

		file := File{}
		err := ds.Get(c, fileKey, &file)
		if err != nil {
			if verr.Root(err) != datastore.ErrNoSuchEntity {
				return err
			}

			file.Key = fileKey
		}

		file.Size = int64(len(data))
		err = ds.Put(ctx, &file)
		if err != nil {
			return err
		}

		fileData := &FileData{
			Content: data,
			Mime:    mime,
		}
		fileData.SetKey(createFileDataKey(r, id))
		err = ds.Put(c, fileData)
		if err != nil {
			return err
		}
		return nil
	}, option)
	return nil
}

func ExistFile(r *http.Request, id string) bool {

	c := appengine.NewContext(r)
	file := &File{}
	file.Key = createFileKey(r, id)
	err := ds.Get(c, file.Key, file)
	if err != nil {
		if verr.Root(err) != datastore.ErrNoSuchEntity {
			return true
		} else if verr.Root(err) == datastore.ErrNoSuchEntity {
			return false
		}
	}
	return true
}

func RemoveFile(r *http.Request, id string) error {
	c := appengine.NewContext(r)

	option := &datastore.TransactionOptions{XG: true}
	err := datastore.RunInTransaction(c, func(ctx context.Context) error {
		fkey := createFileKey(r, id)
		err := ds.Delete(c, fkey)
		if err != nil {
			return err
		}
		fdkey := createFileDataKey(r, id)
		return ds.Delete(c, fdkey)
	}, option)
	return err
}

const KindFileDataName = "FileData"

type FileData struct {
	key     *datastore.Key
	Mime    string
	Content []byte
}

func (d *FileData) GetKey() *datastore.Key {
	return d.key
}

func (d *FileData) SetKey(k *datastore.Key) {
	d.key = k
}

func createFileDataKey(r *http.Request, name string) *datastore.Key {
	c := appengine.NewContext(r)
	return datastore.NewKey(c, KindFileDataName, name, 0, nil)
}

func convertImage(r io.Reader) ([]byte, bool, error) {

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, false, err
	}

	var img image.Image
	buff := bytes.NewBuffer(b)
	cnv := false
	//over 1mb
	if len(b) > (1 * 1024 * 1024) {
		if img == nil {
			img, _, err = image.Decode(buff)
			if err != nil {
				return nil, false, err
			}
		}

		img = resize.Resize(1000, 0, img, resize.Lanczos3)
		cnv = true
	}

	if cnv {
		buffer := new(bytes.Buffer)
		if err := jpeg.Encode(buffer, img, nil); err != nil {
			return nil, cnv, err
		}
		b = buffer.Bytes()
	}

	return b, cnv, nil
}

func SelectFileData(r *http.Request, name string) (*FileData, error) {
	c := appengine.NewContext(r)

	rtn := FileData{}
	key := createFileDataKey(r, name)

	err := ds.Get(c, key, &rtn)
	if err != nil {
		if verr.Root(err) != datastore.ErrNoSuchEntity {
			return nil, verr.Root(err)
		} else if verr.Root(err) == datastore.ErrNoSuchEntity {
			return nil, nil
		}
	}
	return &rtn, nil
}
