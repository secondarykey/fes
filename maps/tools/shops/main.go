// 店舗データ（JSON）の生成
//
// マップ画像の店舗一覧を書き起こした対応表（TSV）と、出店申込フォームの回答（CSV）から
// アプリが読み込む店舗データを作る。申込には電話番号や住所が含まれるため、
// 申込の CSV はリポジトリに置かず、出力にも店名・説明・URL 以外は含めない。
//
//	cd maps
//	go run tools/shops/main.go -apply "<申込の CSV>" -list data/2026-autumn/shops.tsv -out src/data/shops.json
//
// 対応表の列
//
//	key    店舗番号（マップ上の番号。画像ファイル名にも使う）
//	group  food / kitchencar / goods
//	name   表示する店名
//	desc   マップ画像の店舗一覧の説明
//	app    申込の店名（完全一致）。申込が無ければ空
//	text   説明文。空なら申込の販売メニュー、"-" なら desc、それ以外はその文（\n で改行）
//	links  リンク。空なら申込の URL 欄から作る、"-" なら無し、それ以外は空白区切りの URL
package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// 申込 CSV の列
const (
	colName  = 1
	colMenu  = 4
	colURL   = 8
	colCount = 14
)

type Shop struct {
	Key    string   `json:"key"`
	Group  string   `json:"group"`
	Name   string   `json:"name"`
	Detail string   `json:"detail"`
	Links  []string `json:"links"`
}

// 申込の内容（同じ店名で複数回答があれば、後の回答の空でない値で上書きする）
type application struct {
	menu string
	url  string
}

func main() {
	apply := flag.String("apply", "", "出店申込の回答 CSV")
	list := flag.String("list", "", "店舗の対応表 TSV")
	out := flag.String("out", "", "出力する JSON")
	flag.Parse()

	if *apply == "" || *list == "" || *out == "" {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(*apply, *list, *out); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

func run(apply, list, out string) error {

	apps, err := readApplications(apply)
	if err != nil {
		return fmt.Errorf("申込の読み込み: %w", err)
	}

	rows, err := readList(list)
	if err != nil {
		return fmt.Errorf("対応表の読み込み: %w", err)
	}

	used := make(map[string]bool)
	shops := make([]Shop, 0, len(rows))
	for _, r := range rows {
		s, err := build(r, apps)
		if err != nil {
			return fmt.Errorf("%s %s: %w", r["key"], r["name"], err)
		}
		used[r["app"]] = true
		shops = append(shops, s)
	}

	for name := range apps {
		if !used[name] {
			fmt.Fprintf(os.Stderr, "対応表に無い申込: %s\n", name)
		}
	}

	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(shops)
}

func build(r map[string]string, apps map[string]application) (Shop, error) {

	s := Shop{Key: r["key"], Group: r["group"], Name: r["name"], Links: []string{}}

	var app application
	if r["app"] != "" {
		var ok bool
		app, ok = apps[r["app"]]
		if !ok {
			return s, fmt.Errorf("申込が見つからない: %s", r["app"])
		}
	}

	switch text := r["text"]; {
	case text == "-":
		s.Detail = r["desc"]
	case text != "":
		s.Detail = strings.ReplaceAll(text, `\n`, "\n")
	case app.menu != "":
		s.Detail = cleanText(app.menu)
	default:
		s.Detail = r["desc"]
	}

	switch links := r["links"]; {
	case links == "-":
	case links != "":
		s.Links = strings.Fields(links)
	default:
		s.Links = parseLinks(app.url, func(token string) {
			fmt.Fprintf(os.Stderr, "%s %s: URL にできない: %s\n", s.Key, s.Name, token)
		})
		if len(s.Links) > 1 {
			fmt.Fprintf(os.Stderr, "%s %s: リンクが複数: %v\n", s.Key, s.Name, s.Links)
		}
	}
	return s, nil
}

func readApplications(name string) (map[string]application, error) {

	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = colCount

	apps := make(map[string]application)
	first := true
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if first {
			first = false
			continue
		}

		key := strings.TrimSpace(rec[colName])
		app := apps[key]
		if v := strings.TrimSpace(rec[colMenu]); v != "" {
			app.menu = v
		}
		if v := strings.TrimSpace(rec[colURL]); v != "" {
			app.url = v
		}
		apps[key] = app
	}
	return apps, nil
}

func readList(name string) ([]map[string]string, error) {

	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	header := strings.Split(lines[0], "\t")

	var rows []map[string]string
	for i, line := range lines[1:] {
		if line == "" {
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) != len(header) {
			return nil, fmt.Errorf("%d 行目の列数が %d（%d 必要）", i+2, len(cols), len(header))
		}
		row := make(map[string]string, len(header))
		for j, h := range header {
			row[h] = strings.TrimSpace(cols[j])
		}
		rows = append(rows, row)
	}
	return rows, nil
}

var blankLines = regexp.MustCompile(`\n{3,}`)

// 申込の文を表示用に整える（行頭・行末の空白とタブ、連続する空行を除く）
func cleanText(s string) string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSpace(strings.ReplaceAll(l, "\t", " "))
	}
	return blankLines.ReplaceAllString(strings.Join(lines, "\n"), "\n\n")
}

var domain = regexp.MustCompile(`^[\w.-]+\.(com|jp|net|org|fm|tokyo)(/\S*)?$`)

// 申込の URL 欄からリンクを作る
// "@xxx" は Instagram のアカウントとみなす。判断できない語は unknown に渡す
func parseLinks(s string, unknown func(string)) []string {

	links := []string{}
	for _, token := range strings.Fields(s) {
		var link string
		switch {
		case strings.HasPrefix(token, "http://") || strings.HasPrefix(token, "https://"):
			link = normalizeURL(token)
		case strings.HasPrefix(token, "@") && len(token) > 1:
			link = "https://www.instagram.com/" + token[1:] + "/"
		case domain.MatchString(token):
			link = "https://" + token
		default:
			if strings.ContainsAny(token, "._") {
				unknown(token)
			}
			continue
		}
		links = append(links, link)
	}
	return links
}

// Instagram のリンクから共有用のパラメータを除き、"/@xxx" を "/xxx" にする
func normalizeURL(s string) string {
	u, err := url.Parse(s)
	if err != nil || !strings.HasSuffix(u.Host, "instagram.com") {
		return s
	}
	u.RawQuery = ""
	u.Scheme = "https"
	u.Host = "www.instagram.com"
	u.Path = strings.Replace(u.Path, "/@", "/", 1)
	return u.String()
}
