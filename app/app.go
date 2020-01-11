package app

import (
	"app/handler"

	"google.golang.org/appengine"
)

func Start() error {

	handler.Register()
	appengine.Main()
	return nil
}
