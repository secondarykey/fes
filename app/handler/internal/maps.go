package internal

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed all:_assets/maps/*
var embMaps embed.FS
var mapsFs fs.FS

func init() {
	var err error
	mapsFs, err = fs.Sub(embMaps, "_assets/maps")
	if err != nil {
		log.Printf("%+v", err)
	}
}

func RegisterMaps(mux *http.ServeMux) error {
	fs := http.FileServer(http.FS(mapsFs))
	mux.Handle("/maps/", http.StripPrefix("/maps/", fs))
	return nil
}
