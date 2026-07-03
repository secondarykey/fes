package app

import (
	"app/config"
	"app/handler"
	"app/handler/manage"
	"fmt"
	"net/http"

	"golang.org/x/xerrors"
)

func Listen(opts ...config.Option) error {

	err := config.Set(opts)
	if err != nil {
		return xerrors.Errorf("config.Set() error: %w", err)
	}

	mux, err := registerHandler()
	if err != nil {
		return xerrors.Errorf("registerHandler() error: %w", err)
	}

	conf := config.Get()
	serve := fmt.Sprintf(":%d", conf.Port)

	fmt.Printf("Fes Start! Listen[%s]\n", serve)
	err = http.ListenAndServe(serve, mux)
	if err != nil {
		return xerrors.Errorf("http.ListenAndServe error: %w", err)
	}
	return nil
}

// registerHandler は全ハンドラを登録した ServeMux を返す。
// DefaultServeMux は使用しない（pprof 等が意図せず公開されるのを防ぐ）。
func registerHandler() (*http.ServeMux, error) {
	mux := http.NewServeMux()
	err := manage.Register(mux)
	if err != nil {
		return nil, xerrors.Errorf("manage handler register error: %w", err)
	}
	err = handler.Register(mux)
	if err != nil {
		return nil, xerrors.Errorf("handler register error: %w", err)
	}
	return mux, nil
}
