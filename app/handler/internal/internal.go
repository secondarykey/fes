package internal

import (
	"fmt"
	"net/http"
)

type CacheServer struct {
	age     int
	handler http.Handler
}

func (cs *CacheServer) SetFS(f http.FileSystem, s int) {
	cs.age = s
	cs.handler = http.FileServer(f)
}

func (cs CacheServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if cs.age > 0 {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", cs.age))
	}
	cs.handler.ServeHTTP(w, r)
}

func GrantFS(f http.FileSystem, g string) *grantFS {
	var grant grantFS
	grant.fs = f
	grant.prefix = g
	return &grant
}

type grantFS struct {
	fs     http.FileSystem
	prefix string
}

func (g grantFS) Open(name string) (http.File, error) {
	n := g.prefix + name
	return g.fs.Open(n)
}
