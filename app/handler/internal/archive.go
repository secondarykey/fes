package internal

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
)

//go:embed all:_assets/archives
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
	archiveHandle.SetFS(http.FS(archiveFs), 86400)

	for _, dir := range dirs {
		fmt.Printf("Archives[%s]\n", dir)
		url := fmt.Sprintf("/%s/", dir)
		http.Handle(url, archiveHandle)
	}
	return nil
}
