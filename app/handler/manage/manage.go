package manage

import (
	"app/datastore"
	"app/logic"

	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/http/pprof"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
)

func Register(root *http.ServeMux) error {

	if err := InitSessionStore(); err != nil {
		return xerrors.Errorf("InitSessionStore() error: %w", err)
	}

	r := mux.NewRouter()
	s := r.PathPrefix("/manage").Subrouter()

	// ページプレビュー（/manage/page/view/）— SPA catch-all より前に登録
	s.HandleFunc("/page/view/{key}", privatePageHandler).Methods("GET")
	s.HandleFunc("/page/view/", privateHandler).Methods("GET")

	// ページ画像の表示（SPA の <img> から参照される）
	s.HandleFunc("/file/view/{key}", fileViewHandler).Methods("GET")

	// pprof（/manage/ 配下のためログイン必須）— SPA catch-all より前に登録
	d := s.PathPrefix("/debug/pprof").Subrouter()
	d.HandleFunc("/cmdline", pprof.Cmdline)
	d.HandleFunc("/profile", pprof.Profile)
	d.HandleFunc("/symbol", pprof.Symbol)
	d.HandleFunc("/trace", pprof.Trace)
	d.PathPrefix("/").Handler(http.StripPrefix("/manage", http.HandlerFunc(pprof.Index)))

	// SPA + API (/manage/api/ + /manage/ キャッチオール)
	if err := registerAPI(s, root); err != nil {
		return xerrors.Errorf("registerAPI() error: %w", err)
	}

	h := NewHandler(s)
	root.Handle("/manage/", h)

	return nil
}

type ManageHandler struct {
	r *mux.Router
}

func NewHandler(r *mux.Router) ManageHandler {
	return ManageHandler{r}
}

func (h ManageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//CSRF 対策: 状態変更リクエストは同一オリジンからのみ受け付ける
	if !isSafeMethod(r.Method) && !isSameOrigin(r) {
		log.Printf("cross-origin request rejected: %s %s (Origin=%q Referer=%q)",
			r.Method, r.URL.Path, r.Header.Get("Origin"), r.Header.Get("Referer"))
		if isAPIPath(r.URL.Path) {
			apiError(w, "forbidden: cross-origin request", http.StatusForbidden)
			return
		}
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	//セッションの存在を確認
	u, err := GetSession(r)
	if err != nil {
		log.Printf("%+v", err)
		if isAPIPath(r.URL.Path) {
			apiUnauthorized(w)
			return
		}
		http.Redirect(w, r, "/login?redirect="+r.URL.RequestURI(), 302)
		return
	}

	if u == nil {
		log.Println("ユーザがいない")
		if isAPIPath(r.URL.Path) {
			apiUnauthorized(w)
			return
		}
		http.Redirect(w, r, "/login?redirect="+r.URL.RequestURI(), 302)
		return
	}

	h.r.ServeHTTP(w, r)
}

func isAPIPath(path string) bool {
	return strings.HasPrefix(path, "/manage/api/")
}

func isSafeMethod(m string) bool {
	switch m {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	return false
}

// isSameOrigin は Origin ヘッダ（無ければ Referer）のホストがリクエストの
// Host と一致するかを返す。どちらも無い場合は拒否する。
// ブラウザの fetch / フォーム送信は POST 時に必ず Origin を付けるため、
// 正規の SPA からのリクエストが弾かれることはない。
func isSameOrigin(r *http.Request) bool {
	src := r.Header.Get("Origin")
	if src == "" {
		src = r.Header.Get("Referer")
	}
	if src == "" {
		return false
	}
	u, err := url.Parse(src)
	if err != nil {
		return false
	}
	return u.Host == r.Host
}

func apiUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
}

func privateHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	site, err := dao.SelectSite(ctx, -1)
	if err != nil {
		if err == datastore.SiteNotFoundError {
			errorPage(w, "Not Found", fmt.Errorf("サイトが設定されていません。管理画面の Site Settings から設定してください。"), 404)
			return
		}
		errorPage(w, "Not Found", fmt.Errorf("Root page not found"), 404)
		return
	}
	pageView(w, r, site.Root)
}

func privatePageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]
	pageView(w, r, id)
}

func pageView(w http.ResponseWriter, r *http.Request, id string) {

	page := 1
	val := r.URL.Query()
	pageVal := val.Get("page")
	if pageVal != "" {
		p, err := strconv.Atoi(pageVal)
		if err == nil {
			page = p
		}
	}

	//管理用の書き出し
	err := logic.WriteManageHTML(w, r, id, page, nil)
	if err != nil {
		errorPage(w, "ERROR:Generate HTML", err, 500)
		return
	}
}

func errorPage(w http.ResponseWriter, t string, e error, num int) {

	desc := fmt.Sprintf("%+v", e)
	log.Println(desc)

	w.WriteHeader(num)
	fmt.Fprintf(w,
		"<html><head><title>%s</title></head><body><h1>%s</h1><pre>%s</pre></body></html>",
		html.EscapeString(t), html.EscapeString(t), html.EscapeString(desc))
}
