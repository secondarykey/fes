package main

import (
	"app/config"
	"app/datastore"
	"app/logic"
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func confirm(prompt string) bool {
	fmt.Print(prompt)
	s, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(strings.ToLower(s)) == "y"
}

// このコマンドはローカルのデータから静的なサイトを出力し、GCS にアップロードします。
//
//	使い方: archive [options] <ディレクトリ名>
//	-upload      ディレクトリが存在する場合は生成をスキップして GCS にアップロード
//	-local       ディレクトリが存在する場合は生成をスキップしてローカルサーバで確認
//
// 実行前に gcloud auth application-default login を済ませておくこと（-upload 時）。
func main() {
	uploadFlag := flag.Bool("upload", false, "GCSにアップロード")
	localFlag := flag.Bool("local", false, "ローカルサーバで確認 (http://localhost:3000)")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: archive [options] <ディレクトリ名>")
		flag.Usage()
		os.Exit(1)
	}
	dir := flag.Arg(0)

	err := config.Set([]config.Option{
		config.SetProjectID(),
		config.SetDatastore(),
	})
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// ディレクトリが存在するか確認
	dirExists := false
	if _, err := os.Stat(dir); err == nil {
		dirExists = true
	}

	// フラグがある場合かつディレクトリ存在時は生成をスキップ
	skipCreate := (*uploadFlag || *localFlag) && dirExists

	if !skipCreate {
		fmt.Println("=== Creating static site:", dir)
		if err := logic.CreateStaticSite(dir); err != nil {
			log.Fatalf("%+v", err)
		}
	} else {
		fmt.Printf("=== Skip creation (directory already exists): %s\n", dir)
	}

	switch {
	case *localFlag:
		prefix := "/" + dir + "/"
		http.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(dir))))
		fmt.Println("=== Local server -> http://localhost:3000" + prefix)
		if err := http.ListenAndServe(":3000", nil); err != nil {
			log.Fatalf("%+v", err)
		}

	default: // -upload または引数なし
		bucket := resolveArchiveBucket()
		fmt.Printf("\nUpload target:\n  directory : %s\n  bucket    : %s\n", dir, bucket)
		if !confirm("Upload? [y/N]: ") {
			fmt.Println("Cancelled.")
			return
		}
		fmt.Println("=== Uploading to GCS bucket:", bucket)
		if err := logic.UploadToGCS(dir, bucket); err != nil {
			log.Fatalf("%+v", err)
		}
		fmt.Println("=== Done!")
	}
}

func resolveArchiveBucket() string {
	ctx := context.Background()
	dao := datastore.NewDao()
	defer dao.Close()

	site, err := dao.SelectSite(ctx, -1)
	if err != nil {
		log.Printf("Warning: could not read Site from Datastore: %v", err)
		log.Printf("Using default bucket: %s", config.ArchiveBucket)
		return config.ArchiveBucket
	}

	if site.ArchiveBucket != "" {
		return site.ArchiveBucket
	}
	return config.ArchiveBucket
}
