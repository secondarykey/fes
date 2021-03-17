package internal

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
)

//go:embed _assets/archives
var embArchive embed.FS
var archiveFs fs.FS

func init() {
	var err error
	archiveFs, err = fs.Sub(embArchive, "_assets/archives")
	if err != nil {
		log.Printf("%+v", err)
	}
}

func RegisterArchive(dirs ...string) error {

	var archiveHandle CacheServer
	archiveHandle.SetFS(archiveFs, 86400)

	for _, dir := range dirs {
		url := fmt.Sprintf("/%s/", dir)
		http.Handle(url, archiveHandle)
	}
	return nil
}
