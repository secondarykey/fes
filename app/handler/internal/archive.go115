package internal

import (
	"fmt"
	"net/http"
)

func RegisterArchive(dirs ...string) error {

	var archiveHandle CacheServer
	archiveHandle.SetFS(GrantFS(statikFS, "/archives"), 86400)
	for _, dir := range dirs {
		url := fmt.Sprintf("/%s/", dir)
		http.Handle(url, archiveHandle)
	}
	return nil
}
