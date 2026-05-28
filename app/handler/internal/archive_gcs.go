package internal

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/xerrors"
)

func RegisterGCSArchive(bucket string, names []string) error {
	if len(names) == 0 {
		return nil
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return xerrors.Errorf("storage.NewClient() error: %w", err)
	}

	for _, name := range names {
		pre := "/" + name + "/"
		http.Handle(pre, http.StripPrefix(pre, gcsArchiveHandler(client, bucket, name+"/")))
		fmt.Printf("Archive(GCS):[%s] -> gs://%s/%s/\n", pre, bucket, name)
	}

	return nil
}

func gcsArchiveHandler(client *storage.Client, bucket, prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" || strings.HasSuffix(path, "/") {
			path += "index.html"
		}

		objName := prefix + path
		rc, err := client.Bucket(bucket).Object(objName).NewReader(r.Context())
		if err == storage.ErrObjectNotExist {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			log.Printf("gcsArchiveHandler NewReader(%s) error: %v", objName, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer rc.Close()

		lastModified := rc.Attrs.LastModified
		w.Header().Set("Content-Type", rc.Attrs.ContentType)
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Set("Last-Modified", lastModified.UTC().Format(http.TimeFormat))

		if ims := r.Header.Get("If-Modified-Since"); ims != "" {
			if t, err := http.ParseTime(ims); err == nil {
				if !lastModified.Truncate(time.Second).After(t) {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
		}

		if _, err := io.Copy(w, rc); err != nil {
			log.Printf("gcsArchiveHandler Copy(%s) error: %v", objName, err)
		}
	})
}
