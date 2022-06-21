package handler

import (
	"app/datastore"
	. "app/handler/internal"
	"app/logic"

	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
)

func init() {
	setEnvironment()
}

func setEnvironment() {

	m := GetEnvironmentMap()

	if m == nil {
		log.Println("GetEnvironmentMap() is nil")
		return
	}

	for k, v := range m {
		err := os.Setenv(k, v)
		if err != nil {
			log.Println("os.Setenv() error: %v", err)
		}
	}
}

func Register() error {

	//TODO
	// アーカイブの自動化

	err := RegisterArchive("2020", "2021", "2022-Spring")
	if err != nil {
		return xerrors.Errorf("RegisterArchive() error: %w", err)
	}

	//外部アクセス
	r := mux.NewRouter()

	err = RegisterStatic()
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

	http.Handle("/", r)

	return nil
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

func favicon(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "public, max-age=86400")

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	//ファイルが存在するか？
	fav, err := dao.GetFavicon(ctx)
	if err != nil {
		log.Printf("GetFavicon error: %+v", err)
		return
	}

	if fav != nil {
		w.Write(fav)
		return
	}

	b, err := GetSystemFavicon()
	if err != nil {
		log.Printf("GetSystemFavicon error: %+v", err)
		return
	}
	w.Write(b)
}

func robotTxt(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "public, max-age=86400")

	host := getHost(r)

	txt := fmt.Sprintf(`User-agent:*
Disallow:/file/*
Disallow:/manage/
Sitemap:%ssitemap/
Sitemap:%ssitemap.xml`, host, host)
	w.Write([]byte(txt))
}

func sitemap(w http.ResponseWriter, r *http.Request) {
	root := getHost(r)
	// 60 * 60 * 24
	w.Header().Set("Cache-Control", "public, max-age=86400")
	err := GenerateSitemap(r.Context(), root, w)
	if err != nil {
		errorPage(w, r, "Generate sitemap error", err, 500)
	}
}

func getHost(r *http.Request) string {
	scheme := "http"
	if strings.Index(r.Host, "localhost") == -1 {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/", scheme, r.Host)
}
