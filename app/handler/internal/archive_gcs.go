package internal

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/xerrors"
)

const DefaultArchiveDailyLimit = 2000

// GCSArchiveRouter は動的にアーカイブ名を切り替え可能なルーターを提供する。
// 起動時に InitGCSArchive() で初期化し、Update() で設定を変更できる。
var GCSArchiveRouter *gcsArchiveRouter

type gcsArchiveRouter struct {
	mu     sync.RWMutex
	client *storage.Client
	bucket string
	names  map[string]bool

	// アクセスカウンタ
	count    atomic.Int64
	resetAt  atomic.Value // time.Time: 次のリセット時刻
	limit    atomic.Int64
	location *time.Location
}

// InitGCSArchive は GCS クライアントを作成し、初期設定でルーターを構築する。
func InitGCSArchive(bucket string, names []string) (http.Handler, error) {
	if GCSArchiveRouter != nil {
		GCSArchiveRouter.Update(bucket, names)
		return GCSArchiveRouter, nil
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, xerrors.Errorf("storage.NewClient() error: %w", err)
	}

	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}

	loc, _ := time.LoadLocation("Asia/Tokyo")
	if loc == nil {
		loc = time.UTC
	}

	GCSArchiveRouter = &gcsArchiveRouter{
		client:   client,
		bucket:   bucket,
		names:    nameSet,
		location: loc,
	}
	GCSArchiveRouter.limit.Store(DefaultArchiveDailyLimit)
	GCSArchiveRouter.resetAt.Store(nextMidnight(loc))

	for _, name := range names {
		fmt.Printf("Archive(GCS):[/%s/] -> gs://%s/%s/\n", name, bucket, name)
	}

	return GCSArchiveRouter, nil
}

// nextMidnight は次の日本時間 0:00 を返す。
func nextMidnight(loc *time.Location) time.Time {
	now := time.Now().In(loc)
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
	return next
}

// Update はバケット名とアーカイブ名一覧を動的に更新する。
func (g *gcsArchiveRouter) Update(bucket string, names []string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.bucket = bucket
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}
	g.names = nameSet

	log.Printf("GCSArchiveRouter updated: bucket=%s, names=%v", bucket, names)
}

// SetLimit は1日あたりのアクセス上限を設定する。
func (g *gcsArchiveRouter) SetLimit(n int64) {
	g.limit.Store(n)
}

// AccessInfo はアクセスカウンタの現在の状態を返す。
type AccessInfo struct {
	Count   int64     `json:"count"`
	Limit   int64     `json:"limit"`
	ResetAt time.Time `json:"resetAt"`
}

// GetAccessInfo はアクセスカウンタの現在の状態を返す。
func (g *gcsArchiveRouter) GetAccessInfo() AccessInfo {
	g.checkReset()
	return AccessInfo{
		Count:   g.count.Load(),
		Limit:   g.limit.Load(),
		ResetAt: g.resetAt.Load().(time.Time),
	}
}

// checkReset はリセット時刻を過ぎていればカウンタをリセットする。
func (g *gcsArchiveRouter) checkReset() {
	resetAt := g.resetAt.Load().(time.Time)
	if time.Now().Before(resetAt) {
		return
	}
	// リセット実行
	g.count.Store(0)
	g.resetAt.Store(nextMidnight(g.location))
}

// Match はパスがアーカイブ名に一致するかを返す。
func (g *gcsArchiveRouter) Match(path string) bool {
	p := strings.TrimPrefix(path, "/")
	idx := strings.Index(p, "/")
	if idx < 0 {
		return false
	}
	name := p[:idx]

	g.mu.RLock()
	ok := g.names[name]
	g.mu.RUnlock()
	return ok
}

func (g *gcsArchiveRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// URL から /{name}/... を抽出
	p := strings.TrimPrefix(r.URL.Path, "/")
	idx := strings.Index(p, "/")
	if idx < 0 {
		http.NotFound(w, r)
		return
	}
	name := p[:idx]
	rest := p[idx+1:]

	g.mu.RLock()
	ok := g.names[name]
	bucket := g.bucket
	g.mu.RUnlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	// アクセスカウンタのリセット確認とインクリメント
	g.checkReset()
	current := g.count.Add(1)
	limit := g.limit.Load()
	if limit > 0 && current > limit {
		w.Header().Set("Retry-After", "86400")
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	if rest == "" || strings.HasSuffix(rest, "/") {
		rest += "index.html"
	}

	objName := name + "/" + rest
	rc, err := g.client.Bucket(bucket).Object(objName).NewReader(r.Context())
	if err == storage.ErrObjectNotExist {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("gcsArchiveRouter NewReader(%s) error: %v", objName, err)
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
		log.Printf("gcsArchiveRouter Copy(%s) error: %v", objName, err)
	}
}
