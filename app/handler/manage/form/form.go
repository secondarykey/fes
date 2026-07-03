package form

import (
	"app/datastore"
	"app/logic"

	"errors"
	"net/http"

	"golang.org/x/xerrors"
)

func SetFile(r *http.Request, fs *datastore.FileSet, ft int) error {

	upload, header, err := r.FormFile("file")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			return xerrors.Errorf("FromFile() error: %w", err)
		} else {
			return nil
		}
	}
	//ファイルデータの作成
	defer upload.Close()

	b, flg, err := logic.ConvertImage(upload)
	if err != nil {
		return xerrors.Errorf("convertImage() error: %w", err)
	}

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

	return nil
}
