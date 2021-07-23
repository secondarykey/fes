package app

import (
	"app/config"
	"app/logic"
	"fmt"
	"net/http"

	"golang.org/x/xerrors"
)

//アーカイブ作成用
func CreateStaticSite(dir string, opts ...config.Option) error {

	err := config.Set(opts)
	if err != nil {
		return xerrors.Errorf("config.Set() error: %w", err)
	}

	err = logic.CreateStaticSite(dir)
	if err != nil {
		return xerrors.Errorf("logic.StaticSite() error: %w", err)
	}

	//TODO 起動するかしないかを設定

	prefix := "/" + dir + "/"
	http.Handle(prefix,
		http.StripPrefix(prefix, http.FileServer(http.Dir(dir))))

	fmt.Println("CheckHTTP Server -> ", "http://localhost:3000"+prefix)

	err = http.ListenAndServe(":3000", nil)
	if err != nil {
		return xerrors.Errorf("http.ListenAndServe error: %w")
	}
	return nil
}

func GenerateFiles(dir string, opts ...config.Option) error {

	err := config.Set(opts)
	if err != nil {
		return xerrors.Errorf("config.Set() error: %w", err)
	}

	err = logic.GenerateFiles(dir)
	if err != nil {
		return xerrors.Errorf("logic.GenerateFiles() error: %w", err)
	}
	return nil
}
