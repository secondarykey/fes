package internal

import (
	"errors"
	"net/http"

	"app/datastore"
	"app/logic"

	"golang.org/x/xerrors"
)

func CreateFormFile(r *http.Request, ft int) (*datastore.FileSet, error) {

	upload, header, err := r.FormFile("file")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			return nil, xerrors.Errorf("FromFile() error: %w", err)
		} else {
			return nil, nil
		}
	}
	//ファイルデータの作成
	defer upload.Close()

	b, flg, err := logic.ConvertImage(upload)
	if err != nil {
		return nil, xerrors.Errorf("convertImage() error: %w", err)
	}

	var fs datastore.FileSet
	var f datastore.File
	var fd datastore.FileData

	fs.Name = header.Filename

	f.Size = int64(len(b))
	f.Type = ft

	mime := header.Header["Content-Type"][0]
	if flg {
		mime = "image/jpeg"
	}
	fd.Content = b
	fd.Mime = mime

	fs.File = &f
	fs.FileData = &fd

	return &fs, nil
}
