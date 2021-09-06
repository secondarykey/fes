package internal

import (
	"embed"
	"io"
	"io/fs"
	"log"
	"net/http"

	"golang.org/x/xerrors"
)

//go:embed _assets/static
var embStatic embed.FS
var manageFs fs.FS
var publicFs fs.FS

func init() {
	var err error
	manageFs, err = fs.Sub(embStatic, "_assets/static/manage")
	if err != nil {
		log.Printf("%+v", err)
	}
	publicFs, err = fs.Sub(embStatic, "_assets/static/images")
	if err != nil {
		log.Printf("%+v", err)
	}
}

func RegisterStatic() error {
	//TODO Deprecated
	fs := http.StripPrefix("/images/", http.FileServer(http.FS(publicFs)))
	http.Handle("/images/", fs)
	return nil
}

func RegisterManageStatic() error {

	fs := http.StripPrefix("/manage/", http.FileServer(http.FS(manageFs)))

	http.Handle("/manage/favicon.ico", fs)
	http.Handle("/manage/js/", fs)
	http.Handle("/manage/css/", fs)

	return nil
}

func GetSystemFavicon() ([]byte, error) {

	fp, err := manageFs.Open("favicon.ico")
	if err != nil {
		return nil, xerrors.Errorf("manageFs.Open() error: %w", err)
	}
	defer fp.Close()

	b, err := io.ReadAll(fp)
	if err != nil {
		return nil, xerrors.Errorf("io.ReadAll() error: %w", err)
	}
	return b, nil
}
