package manage

import (
	"app/datastore"

	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"
)

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	err := datastore.RefreshSite(r.Context())
	if err != nil {
		errorPage(w, "Refresh Site Error", err, 500)
		return
	}
	w.Write([]byte("Success."))
}

func backupHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	//バイナリを作成
	data, err := datastore.GetBackupData(ctx)
	if err != nil {
		errorPage(w, "Create BackupDataError", err, 500)
		return
	}

	//Writeでコピーする
	w.Header().Add("Content-Type", "application/zip")
	now := time.Now()
	w.Header().Set("Content-Disposition", "attachment; filename=fes-backup-"+now.Format("20060102150405")+".zip")
	//Zipにする
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for kind, elm := range data {

		for key, byt := range elm {

			esp := strings.Replace(key, "?", "_QUESTION_", -1)
			esp = strings.Replace(esp, "=", "_EQUAL_", -1)

			writer, err := zipWriter.Create(kind + "/" + esp)
			if err != nil {
				errorPage(w, "Create Zip", err, 500)
				return
			}
			_, err = writer.Write(byt)
			if err != nil {
				errorPage(w, "Write Zip", err, 500)
				return
			}
		}
	}

}

func restoreHandler(w http.ResponseWriter, r *http.Request) {

	file, header, err := r.FormFile("restoreFile")
	if err != nil {
		errorPage(w, "Read Zip", err, 500)
		return
	}
	defer file.Close()

	//ZIPを解析
	reader, err := zip.NewReader(file, header.Size)
	if err != nil {
		errorPage(w, "Read Error", err, 500)
		return
	}

	backup, err := createGob(reader)
	if err != nil {
		errorPage(w, "CreateGob Error", err, 500)
		return
	}

	ctx := r.Context()
	//Putする
	err = datastore.PutBackupData(ctx, backup)
	if err != nil {
		errorPage(w, "Put Error", err, 500)
		return
	}

	//TODO redirect???
	viewSiteHandler(w, r)
}

func createGob(closer *zip.Reader) (datastore.BackupData, error) {

	rtn := make(datastore.BackupData)
	for _, elm := range closer.File {

		name := elm.Name
		fileReader, err := elm.Open()
		if err != nil {
			return nil, err
		}

		nameArray := strings.Split(name, "/")
		//Fileをパスにしたら駄目
		kind := nameArray[0]
		key := nameArray[1]

		writer := bytes.NewBuffer(nil)
		_, err = io.Copy(writer, fileReader)
		if err != nil {
			return nil, err
		}

		gob, ok := rtn[kind]
		if !ok {
			gob = make(datastore.GobKind)
		}
		gob[key] = writer.Bytes()

		rtn[kind] = gob
	}
	return rtn, nil
}
