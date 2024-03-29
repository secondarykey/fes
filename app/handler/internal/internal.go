package internal

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

//go:embed _assets/environment.json
var envJson []byte

func GetEnvironmentMap() map[string]string {

	envMap := make(map[string]string)

	err := json.Unmarshal(envJson, &envMap)
	if err != nil {
		log.Println(err)
		return nil
	}

	return envMap
}

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
