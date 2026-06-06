package handler

import (
	"fmt"
	"net/http"
	"time"
)

// calcMaxAge は最終更新時刻からの経過時間に応じた max-age 秒数を返す。
// 1時間以内: 60秒, 1〜24時間: 3600秒, 24時間以降: 86400秒
func calcMaxAge(t time.Time) int {
	since := time.Since(t)
	switch {
	case since < time.Hour:
		return 60
	case since < 24*time.Hour:
		return 3600
	default:
		return 86400
	}
}

func setCacheHeaders(w http.ResponseWriter, t time.Time) {
	if t.IsZero() {
		// 更新日時が不明な場合は毎回再検証させる
		w.Header().Set("Cache-Control", "no-cache")
		return
	}
	maxAge := calcMaxAge(t)
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
	w.Header().Set("Last-Modified", t.UTC().Format(http.TimeFormat))
}

// checkNotModified は If-Modified-Since ヘッダを評価し、
// 304 を返すべき場合は true を返す。
func checkNotModified(r *http.Request, t time.Time) bool {
	if t.IsZero() {
		return false
	}
	ims := r.Header.Get("If-Modified-Since")
	if ims == "" {
		return false
	}
	imsTime, err := http.ParseTime(ims)
	if err != nil {
		return false
	}
	return !t.Truncate(time.Second).After(imsTime)
}
