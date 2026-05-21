package internal

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
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

func newCacheServer(age int, f fs.FS) *CacheServer {
	var cs CacheServer
	cs.age = age
	cs.handler = http.FileServerFS(f)
	return &cs
}

func (cs CacheServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if cs.age > 0 {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", cs.age))
	}
	cs.handler.ServeHTTP(w, r)
}
