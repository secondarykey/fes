package fes

import (
	_ "api"
	_ "datastore"
	_ "manage"

	_ "github.com/gorilla/mux"
	_ "github.com/knightso/base/gae/ds"
	_ "github.com/nfnt/resize"
	_ "github.com/satori/go.uuid"
)
