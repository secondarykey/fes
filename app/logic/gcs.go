package logic

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"golang.org/x/xerrors"
)

var gcsContentTypes = map[string]string{
	".html": "text/html; charset=utf-8",
	".css":  "text/css; charset=utf-8",
	".js":   "application/javascript",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
	".ico":  "image/x-icon",
	".pdf":  "application/pdf",
}

func UploadToGCS(dir, bucket string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return xerrors.Errorf("storage.NewClient() error: %w", err)
	}
	defer client.Close()

	bkt := client.Bucket(bucket)
	count := 0

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		objName := dir + "/" + filepath.ToSlash(rel)

		f, err := os.Open(path)
		if err != nil {
			return xerrors.Errorf("os.Open(%s) error: %w", path, err)
		}
		defer f.Close()

		contentType := gcsContentTypes[filepath.Ext(path)]
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		wc := bkt.Object(objName).NewWriter(ctx)
		wc.ContentType = contentType

		if _, err := io.Copy(wc, f); err != nil {
			wc.Close()
			return xerrors.Errorf("upload copy(%s) error: %w", objName, err)
		}
		if err := wc.Close(); err != nil {
			return xerrors.Errorf("upload close(%s) error: %w", objName, err)
		}

		count++
		fmt.Printf("  uploaded: %s\n", objName)
		return nil
	})
	if err != nil {
		return xerrors.Errorf("WalkDir error: %w", err)
	}

	fmt.Printf("Upload complete: %d files -> gs://%s/%s/\n", count, bucket, dir)
	return nil
}
