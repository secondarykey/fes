package internal

import (
	"fmt"
	"io/fs"
	"net/http"
)

type CacheServer struct {
	age     int
	handler http.Handler
}

func (cs *CacheServer) SetFS(f fs.FS, s int) {
	cs.age = s
	cs.handler = http.FileServer(http.FS(f))
}

func (cs CacheServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if cs.age > 0 {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", cs.age))
	}
	cs.handler.ServeHTTP(w, r)
}
