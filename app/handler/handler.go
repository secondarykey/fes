package handler

import (
	"app/datastore"
	"app/handler/internal"
	"app/logic"

	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
)

func init() {
	setEnvironment()
}

func setEnvironment() {

	m := internal.GetEnvironmentMap()

	if m == nil {
		log.Println("GetEnvironmentMap() is nil")
		return
	}

	for k, v := range m {
		err := os.Setenv(k, v)
		if err != nil {
			log.Printf("os.Setenv() error: %v", err)
		}
	}
}

func Register(root *http.ServeMux) error {

	//20260523 廃止
	/*
		err := RegisterArchive()
		if err != nil {
			return xerrors.Errorf("RegisterArchive() error: %w", err)
		}
	*/

	bucket, names, dailyLimit := loadArchiveConfig()
	archiveHandler, err := internal.InitGCSArchive(bucket, names)
	if err != nil {
		log.Printf("InitGCSArchive() error: %+v", err)
	}
	if internal.GCSArchiveRouter != nil && dailyLimit > 0 {
		internal.GCSArchiveRouter.SetLimit(int64(dailyLimit))
	}

	err = internal.RegisterMaps(root)
	if err != nil {
		return xerrors.Errorf("RegisterMaps() error: %w", err)
	}

	//外部アクセス
	r := mux.NewRouter()

	err = internal.RegisterStatic(root)
	if err != nil {
		return xerrors.Errorf("RegisterStatic() error: %w", err)
	}

	r.HandleFunc("/page/{key}", pageHandler).Methods("GET")
	r.HandleFunc("/file/{key}", fileHandler).Methods("GET")
	r.HandleFunc("/file/{date}/{key}", fileDateCacheHandler).Methods("GET")
	r.HandleFunc("/file_cache/{key}", fileCacheHandler).Methods("GET")

	r.HandleFunc("/login", loginHandler).Methods("GET")
	r.HandleFunc("/logout", logoutHandler).Methods("GET")
	r.HandleFunc("/session", sessionHandler).Methods("POST")

	r.HandleFunc("/sitemap.xml", sitemap).Methods("GET")
	r.HandleFunc("/sitemap/", sitemap).Methods("GET")
	r.HandleFunc("/robots.txt", robotTxt).Methods("GET")
	r.HandleFunc("/favicon.ico", favicon).Methods("GET")
	r.HandleFunc("/", indexHandler).Methods("GET")

	//TODO
	// エラーページ
	// JavaScript 埋め込みモード
	// robot.txt

	if archiveHandler != nil {
		// アーカイブルーターが先に判定し、該当しなければ mux ルーターに委譲
		root.Handle("/", archiveFallback(archiveHandler, r))
	} else {
		root.Handle("/", r)
	}

	return nil
}

// archiveFallback はアーカイブ名に一致するリクエストを archive に、それ以外を fallback に委譲する。
func archiveFallback(archive http.Handler, fallback http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if internal.GCSArchiveRouter != nil && internal.GCSArchiveRouter.Match(r.URL.Path) {
			archive.ServeHTTP(w, r)
			return
		}
		fallback.ServeHTTP(w, r)
	})
}

func loadArchiveConfig() (string, []string, int) {
	ctx := context.Background()
	dao := datastore.NewDao()
	defer dao.Close()

	site, err := dao.SelectSite(ctx, -1)
	if err != nil {
		log.Printf("loadArchiveConfig: SelectSite error: %v", err)
		return "", nil, 0
	}

	return site.ArchiveBucket, site.ArchiveNames, site.ArchiveDailyLimit
}

func errorPage(w http.ResponseWriter, r *http.Request, t string, e error, num int) {

	w.WriteHeader(num)

	msg := fmt.Sprintf("%+v", e)
	log.Println(msg)

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	page, err := dao.GetErrorPage(ctx)
	if err != nil {
		log.Printf("%+v", err)
		solidError(w, t, msg)
		return
	}

	if page == nil {
		solidError(w, t, msg)
		return
	}

	var dto logic.ErrorDto

	dto.Message = t
	dto.Detail = msg
	dto.No = num
	//エラーページを作成
	err = logic.WriteManageHTML(w, r, datastore.ErrorPageID, -1, &dto)
	if err != nil {
		log.Printf("%+v", err)
		solidError(w, t, msg)
		return
	}
}

func solidError(w http.ResponseWriter, title, msg string) {
	log.Println("solidError()")
	htm := fmt.Sprintf("<html><head><title>%s</title></head><body><h1>%s</h1></body>", title, msg)
	w.Write([]byte(htm))
}

// favicon は変更頻度が低いためメモリにキャッシュする
var faviconCache struct {
	sync.Mutex
	data    []byte
	expires time.Time
}

const faviconCacheTTL = 10 * time.Minute

func favicon(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "public, max-age=86400")

	b, err := loadFavicon(r.Context())
	if err != nil {
		log.Printf("loadFavicon error: %+v", err)
		return
	}
	w.Write(b)
}

func loadFavicon(ctx context.Context) ([]byte, error) {

	faviconCache.Lock()
	defer faviconCache.Unlock()

	if faviconCache.data != nil && time.Now().Before(faviconCache.expires) {
		return faviconCache.data, nil
	}

	dao := datastore.NewDao()
	defer dao.Close()

	//ファイルが存在するか？
	b, err := dao.GetFavicon(ctx)
	if err != nil {
		return nil, xerrors.Errorf("GetFavicon() error: %w", err)
	}

	if b == nil {
		b, err = internal.GetSystemFavicon()
		if err != nil {
			return nil, xerrors.Errorf("GetSystemFavicon() error: %w", err)
		}
	}

	faviconCache.data = b
	faviconCache.expires = time.Now().Add(faviconCacheTTL)
	return b, nil
}

func robotTxt(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "public, max-age=86400")

	host := getHost(r)

	_, names, _ := loadArchiveConfig()

	var b strings.Builder
	fmt.Fprintf(&b, "User-agent:*\nDisallow:/file/*\nDisallow:/manage/\n")
	for _, name := range names {
		fmt.Fprintf(&b, "Disallow:/%s/\n", name)
	}
	fmt.Fprintf(&b, "Sitemap:%ssitemap/\nSitemap:%ssitemap.xml", host, host)
	w.Write([]byte(b.String()))
}

func sitemap(w http.ResponseWriter, r *http.Request) {
	root := getHost(r)
	// 60 * 60 * 24
	w.Header().Set("Cache-Control", "public, max-age=86400")
	err := internal.GenerateSitemap(r.Context(), root, w)
	if err != nil {
		errorPage(w, r, "Generate sitemap error", err, 500)
	}
}

// getHost はサイトのベース URL（末尾スラッシュ付き）を返す。
// Site 設定の BaseURL を優先し、未設定時のみ Host ヘッダから組み立てる。
// Host ヘッダはクライアントが自由に指定できるため、公開 URL の生成には
// BaseURL を設定しておくことが望ましい。
func getHost(r *http.Request) string {

	dao := datastore.NewDao()
	defer dao.Close()

	site, err := dao.SelectSite(r.Context(), -1)
	if err == nil && site.BaseURL != "" {
		base := site.BaseURL
		if !strings.HasSuffix(base, "/") {
			base += "/"
		}
		return base
	}

	scheme := "http"
	if strings.Index(r.Host, "localhost") == -1 {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/", scheme, r.Host)
}
