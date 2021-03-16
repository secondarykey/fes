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

	conf := config.Get()

	err = manage.Register()
	if err != nil {
		return xerrors.Errorf("manage handler register error: %w", err)
	}
	err = handler.Register()
	if err != nil {
		return xerrors.Errorf("handler register error: %w", err)
	}

	serve := fmt.Sprintf(":%d", conf.Port)

	fmt.Printf("Fes Start![%s]", serve)
	err = http.ListenAndServe(serve, nil)
	return nil
}
