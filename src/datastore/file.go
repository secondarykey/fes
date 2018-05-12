package datastore

import (
	"io"
	"io/ioutil"
	"net/http"

	_ "image/gif"
	_ "image/png"

	verr "github.com/knightso/base/errors"
	"github.com/knightso/base/gae/ds"
	"golang.org/x/net/context"

	"bytes"
	"github.com/nfnt/resize"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"image"
	"image/jpeg"
)

const KIND_FILE = "File"

type File struct {
	Size int64
	ds.Meta
}

func createFileKey(r *http.Request, name string) *datastore.Key {
	c := appengine.NewContext(r)
	return datastore.NewKey(c, KIND_FILE, name, 0, nil)
}

func SelectFiles(r *http.Request) ([]File, error) {

	c := appengine.NewContext(r)
	q := datastore.NewQuery(KIND_FILE).
		Order("- UpdatedAt")

	var s []File

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
	return s, nil
}

func SaveFile(r *http.Request, id string) error {

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
		file := &File{
			Size: int64(len(b)),
		}
		file.Key = createFileKey(r, id)
		err = ds.Put(ctx, file)
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

const KIND_FILEDATA = "FileData"

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
	return datastore.NewKey(c, KIND_FILEDATA, name, 0, nil)
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
