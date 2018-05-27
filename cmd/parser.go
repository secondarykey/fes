package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

type Root struct {
	url  string
	urls []string
	errors []error
}

// 基準URLからサーバ用のURLを取得し、すべてリクエストする処理です
// クラウドの負荷を純粋にかけたいので、ページに対する処理を厳密化しました。
func main() {

	url := "https://www.hagoromo-shizuoka.com/"
	root ,err := NewRoot(url)
	if err != nil {
		log.Fatal(err)
	}

	//一回排除する処理を入れるかな？

	err = root.loop(30)
	if err != nil {
		log.Fatal(err)
	}
}

func NewRoot(root string) (*Root,error) {
	var err error
	r := &Root {
		url : root,
	}

	r.urls,err = getUrls(root)
	if err != nil {
		return nil,err
	}
	r.errors = make([]error,0)
	return r,nil
}

func (r *Root) request() error {

	eg := errgroup.Group{}
	for _, url := range r.urls {
		u := url
		eg.Go(func() error {
			return request(u,nil)
		})
	}

	if err := eg.Wait(); err != nil {
		r.errors = append(r.errors,err)
	}
	return nil
}

func (r *Root) hasError() bool {
	return len(r.errors) > 0
}

func (r *Root) printError() {
	for _,err := range r.errors {
		fmt.Printf("Error:[%v]\n",err)
	}
	r.errors = make([]error,0)
}

func (r *Root) loop(dur time.Duration) error {

	timestamp("Start")
	t := time.NewTicker(dur * time.Second)

	cnt := 1
	errCnt := 0

	for {
		select {
		case <-t.C:
			timestamp(fmt.Sprintf("Access[%s]", r.url))
			err := r.request()
			if err != nil {
				fmt.Printf("Error[%s]\n", err.Error())
				return err
			} else if r.hasError() {
				r.printError()
				errCnt++
			}
			timestamp(fmt.Sprintf("[%06d/%06d]", errCnt,cnt))
			cnt++
		}
	}
	return nil
}

// Style時にURL指定がないかを取得
// 現状、background:url() 形式のみ対応
// 最後にchangeLocal()で実URLを取得しているけど、
// スタイルシートからの相対パスになるので注意
func getStyleUrl(root, line string) string {

	bg := strings.Index(line, "background")
	if bg == -1 {
		return ""
	}

	line = line[bg:]

	url := strings.Index(line, "url")
	if url == -1 {
		return ""
	}

	line = line[url:]

	left := strings.IndexAny(line, "'\"")
	if left == -1 {
		return ""
	}
	line = line[left+1:]

	right := strings.IndexAny(line, "'\"")
	if right == -1 {
		return ""
	}

	line = line[0:right]
	if dist, flg := changeLocal(root, line); flg {
		return dist
	}

	return ""
}

// スタイルシートのデータからURLを取得
func getStyleUrls(root string, b *bytes.Buffer) ([]string, error) {

	urls := make([]string, 0, 5)

	sc := bufio.NewScanner(b)
	for i := 1; sc.Scan(); i++ {
		if err := sc.Err(); err != nil {
			return nil, err
		}

		line := sc.Text()
		url := getStyleUrl(root, line)
		if url != "" {
			urls = append(urls, url)
		}

	}

	return urls, nil
}

// HTML(URL)から各タグで処理されているURLを取得
// 現状 img:src,script:src,link:hrefで行っている
func getUrls(root string) ([]string, error) {

	doc, err := goquery.NewDocument(root)
	if err != nil {
		return nil, err
	}

	urls := make([]string, 1, 10)
	urls[0] = root

	//img タグからURLを抜き出す
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists {
			if dst, flg := changeLocal(root, src); flg {
				urls = append(urls, dst)
			}
		}
	})
	//scriptタグからURLを抜き出す
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists {
			if dst, flg := changeLocal(root, src); flg {
				urls = append(urls, dst)
			}
		}
	})

	//linkタグからURLを抜き出す
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			if dst, flg := changeLocal(root, href); flg {
				urls = append(urls, dst)
				w := &bytes.Buffer{}

				//スタイルシートを取得する
				err := request(dst, w)
				if err == nil {
					styles, err := getStyleUrls(dst, w)
					if err == nil {
						urls = append(urls, styles...)
					}
				}
			}
		}
	})

	return urls, nil
}

// 同一サーバ上にある時、そのURLを返す
// CSS時は相対パスになるので、気を付けること
func changeLocal(root, val string) (string, bool) {
	localPrefix := []string{".", "/", root}
	for _, prefix := range localPrefix {
		if strings.Index(val, prefix) == 0 {

			if prefix == root {
				return val, true
			}
			if prefix == "/" {
				//three index(https://.../)
				idx := 8
				first := strings.Index(root[idx:], "/")
				erase := root[0 : first+idx]
				return erase + val, true
			} else {
				//erase filename
				last := strings.LastIndex(root, "/")
				erase := root[0 : last+1]
				return erase + val, true
			}

		}
	}
	return "", false
}

// リクエスト処理
// 中身が欲しい場合はWriterを渡す
// -> 初回のHTML時は直接goqureryを呼んでいるので、なんか少し変更したい
func request(url string, w io.Writer) error {

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Error StatusCode[%d][%s]", resp.StatusCode, url)
	}

	if w != nil {
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			return err
		}
	}
	return nil
}

//現在時刻の表示
func timestamp(log string) {
	now := time.Now()
	fmt.Println(convertDate(now),log)
}

//JST変換
func convertDate(t time.Time) string {
	if t.IsZero() {
		return "None"
	}
	jst, _ := time.LoadLocation("Asia/Tokyo")
	jt := t.In(jst)
	return jt.Format("2006/01/02 15:04:05")
}

