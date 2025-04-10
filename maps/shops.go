package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const fn = "hb_2025_spring_shops.csv"

const code = `
p = MapPoint.get(%s,%s);
console.log("KEY=%s",p.x,p.y)
this.shops.push(new Shop("%s","%s",%s%s%s,
            new Rect(p.x,p.y,30,50)));
`

/**
 * 店舗情報からJavaScriptのオブジェクトを生成
 * マップ上の座標は緯度経度あればJS上では実装できるかも
 */
func main() {

	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}
}

func run() error {

	in, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create("shops.csv")
	if err != nil {
		return err
	}
	defer out.Close()

	os.Mkdir("images", 0666)

	scanner := bufio.NewScanner(in)

	first := true

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if first {
			first = false
			continue
		}

		err = write(out, strings.Split(line, ","))
		if err != nil {
			return err
		}
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	return nil
}

func write(w io.Writer, clm []string) error {

	//画像があったら、代表画像を作成
	lat := clm[5]
	lon := clm[6]
	key := clm[3]
	name := clm[1]
	detail := strings.ReplaceAll(clm[2], " ", "\n")

	fmt.Println(name)
	url := clm[7]
	if url != "" {
		detail += "\n\n" + url
	}

	for i := 8; i <= 12; i++ {
		img := strings.Trim(clm[i], " ")
		if img != "" {
			fmt.Println(img)
			download(img, key, i-8)
		}
	}

	fmt.Fprintf(w, code+"\n", lat, lon, key, key, name, "`", detail, "`")
	return nil
}

func download(drive string, key string, idx int) {

	id := drive[strings.Index(drive, "?id=")+4:]
	url := fmt.Sprintf("https://drive.google.com/uc?id=%s", id)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	ext := getExtension(contentType)

	out, err := os.Create("images/" + fmt.Sprintf("%s_%d.%s", key, idx, ext))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return
	}
	return

}

var mimeMap = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/gif":  "gif",
	"image/webp": "webp",
	"image/bmp":  "bmp",
	"image/tiff": "tiff",
}

func getExtension(contentType string) string {
	if ext, exists := mimeMap[contentType]; exists {
		return ext
	}
	return "" // 該当なしの場合
}
