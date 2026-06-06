package internal

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/secondarykey/zipfs"
)

// 2026 ignore  //go:embed all:_assets/archives
var embArchive embed.FS
var archiveFs fs.FS

func init() {
	/*
		var err error
		archiveFs, err = fs.Sub(embArchive, "_assets/archives")
		if err != nil {
			log.Printf("%+v", err)
		}
	*/
}

func RegisterArchive() error {

	entries, err := fs.ReadDir(archiveFs, ".")
	if err != nil {
		return err
	}

	var zips []*zipfs.ZipFS
	for _, entry := range entries {

		if entry.IsDir() {
			continue
		}
		n := entry.Name()
		idx := strings.LastIndex(n, ".zip")
		if idx == -1 {
			continue
		}

		name := n[:idx]
		zfs, err := zipfs.New(archiveFs, n)
		if err != nil {
			return err
		}

		pre := fmt.Sprintf("/%s/", name)

		cs := newCacheServer(86400, zfs)
		http.Handle(pre, http.StripPrefix(pre, cs))

		fmt.Printf("Archive:[%s]\n", pre)
		zips = append(zips, zfs)
	}

	go func() {
		for range time.Tick(2 * time.Minute) {
			for _, f := range zips {
				f.Release()
			}
		}
		runtime.GC()
	}()
	return nil
}
