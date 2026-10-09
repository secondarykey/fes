// 店舗画像の取り出しと配置
//
// 出店申込フォームの写真（Google Drive からフォルダごと ZIP でダウンロードしたもの）を店舗番号ごとに分け、
// webp 化したものをアプリの画像ディレクトリに置く。
//
//	cd maps
//	go run tools/images/main.go extract -zip "<ZIP>" -apply "<申込の CSV>" -list data/2026-autumn/shops.tsv
//	（work/images を webp 化する。例: webp -r work/images）
//	go run tools/images/main.go install
//
// extract
//
//	ZIP の全ファイルを work/drive に取り出し（ZIP 内の長いフォルダ名は外す）、
//	店舗の写真を work/images/{key}/{n}.{拡張子} に置く。n は申込の写真の順番（0 始まり）。
//	ZIP には Drive の ID が無いため、ファイルの更新時刻（アップロード時刻）と申込の回答時刻で対応づける。
//	ZIP の時刻と回答時刻の時差は、写真の枚数が一致する回答が最も多くなるものを選ぶ。
//	既にあるファイルは上書きしない。
//
// install
//
//	work/images/{key}/{n}_{幅}x{高さ}.webp を public/images/{key}/{n}.webp に置く。
//	webp が無い写真（50KiB 未満で変換されなかったもの、元が webp のもの）は元の画像をそのまま置く。
package main

import (
	"archive/zip"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// 申込 CSV の列
const (
	colTime   = 0
	colName   = 1
	colImages = 13
)

// 回答の 何分前までのアップロードをその回答の写真とみなすか
const uploadWindow = 30 * time.Minute

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: images extract|install [flags]")
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "extract":
		err = extract(os.Args[2:])
	case "install":
		err = install(os.Args[2:])
	default:
		err = fmt.Errorf("不明なコマンド: %s", os.Args[1])
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

// 申込の 1 回答
type response struct {
	time   time.Time
	name   string
	images int
	files  []*zip.File
}

func extract(args []string) error {

	fs := flag.NewFlagSet("extract", flag.ExitOnError)
	zipName := fs.String("zip", "", "Drive からダウンロードした ZIP")
	apply := fs.String("apply", "", "出店申込の回答 CSV")
	list := fs.String("list", "", "店舗の対応表 TSV")
	work := fs.String("work", "work", "作業ディレクトリ")
	fs.Parse(args)
	if *zipName == "" || *apply == "" || *list == "" {
		fs.Usage()
		os.Exit(2)
	}

	responses, err := readResponses(*apply)
	if err != nil {
		return fmt.Errorf("申込の読み込み: %w", err)
	}
	keys, err := readKeys(*list)
	if err != nil {
		return fmt.Errorf("対応表の読み込み: %w", err)
	}

	zr, err := zip.OpenReader(*zipName)
	if err != nil {
		return err
	}
	defer zr.Close()

	var files []*zip.File
	for _, f := range zr.File {
		if !f.FileInfo().IsDir() {
			files = append(files, f)
		}
	}

	// 全ファイルを work/drive に取り出す
	drive := filepath.Join(*work, "drive")
	used := make(map[string]bool)
	for _, f := range files {
		name := uniqueName(used, path.Base(f.Name))
		if err := writeFile(f, filepath.Join(drive, name)); err != nil {
			return err
		}
	}
	fmt.Fprintf(os.Stderr, "%d ファイルを %s に取り出した\n", len(files), drive)

	offset := detectOffset(files, responses)
	fmt.Fprintf(os.Stderr, "ZIP の時刻 + %v を回答時刻として対応づける\n", offset)
	assign(files, responses, offset)

	// 店名ごとに最後の回答（写真があるもの）を使う
	latest := make(map[string]*response)
	for _, r := range responses {
		if r.images > 0 {
			latest[r.name] = r
		}
	}

	unmatched := 0
	for _, r := range responses {
		if len(r.files) != r.images {
			fmt.Fprintf(os.Stderr, "枚数が合わない: %s %s（申込 %d 枚 / ZIP %d 枚）\n",
				r.time.Format("2006/01/02 15:04:05"), r.name, r.images, len(r.files))
		}
		unmatched += r.images - len(r.files)
	}

	for _, k := range keys {
		r, ok := latest[k.app]
		if !ok {
			continue
		}
		for n, f := range r.files {
			ext := strings.ToLower(path.Ext(f.Name))
			// webp ツールは拡張子で対象を選ぶ（HEIC は heif として扱われる）
			if ext == ".heic" {
				ext = ".heif"
			}
			dst := filepath.Join(*work, "images", k.key, fmt.Sprintf("%d%s", n, ext))
			if err := writeFile(f, dst); err != nil {
				return err
			}
		}
		fmt.Fprintf(os.Stderr, "%s %s: %d 枚\n", k.key, k.app, len(r.files))
	}
	return nil
}

// 時差を 1 時間刻みで試し、写真の枚数が一致する回答が最も多いものを返す
func detectOffset(files []*zip.File, responses []*response) time.Duration {
	best, bestScore := time.Duration(0), -1
	for h := -24; h <= 24; h++ {
		o := time.Duration(h) * time.Hour
		assign(files, responses, o)
		score := 0
		for _, r := range responses {
			if r.images > 0 && len(r.files) == r.images {
				score++
			}
		}
		if score > bestScore {
			best, bestScore = o, score
		}
	}
	return best
}

// 各ファイルを、アップロード時刻の直後の回答に割り当てる
func assign(files []*zip.File, responses []*response, offset time.Duration) {
	for _, r := range responses {
		r.files = nil
	}
	for _, f := range files {
		t := modified(f).Add(offset)
		for _, r := range responses {
			if !r.time.Before(t) && r.time.Sub(t) <= uploadWindow {
				r.files = append(r.files, f)
				break
			}
		}
	}
	for _, r := range responses {
		sort.SliceStable(r.files, func(i, j int) bool {
			return modified(r.files[i]).Before(modified(r.files[j]))
		})
	}
}

// ZIP の更新時刻（タイムゾーンを持たないので UTC として扱う）
func modified(f *zip.File) time.Time {
	m := f.Modified
	return time.Date(m.Year(), m.Month(), m.Day(), m.Hour(), m.Minute(), m.Second(), 0, time.UTC)
}

var driveLink = regexp.MustCompile(`https://drive\.google\.com/open\?id=\S+`)

func readResponses(name string) ([]*response, error) {

	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	recs, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}

	var rtn []*response
	for _, rec := range recs[1:] {
		t, err := time.Parse("2006/01/02 15:04:05", strings.TrimSpace(rec[colTime]))
		if err != nil {
			return nil, fmt.Errorf("タイムスタンプ: %w", err)
		}
		rtn = append(rtn, &response{
			time:   t,
			name:   strings.TrimSpace(rec[colName]),
			images: len(driveLink.FindAllString(rec[colImages], -1)),
		})
	}
	sort.Slice(rtn, func(i, j int) bool { return rtn[i].time.Before(rtn[j].time) })
	return rtn, nil
}

type shopKey struct {
	key string
	app string
}

// 対応表から key と申込の店名を読む
func readKeys(name string) ([]shopKey, error) {

	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	header := strings.Split(lines[0], "\t")
	idx := make(map[string]int)
	for i, h := range header {
		idx[h] = i
	}

	var keys []shopKey
	for _, line := range lines[1:] {
		cols := strings.Split(line, "\t")
		if len(cols) != len(header) {
			continue
		}
		if app := strings.TrimSpace(cols[idx["app"]]); app != "" {
			keys = append(keys, shopKey{key: cols[idx["key"]], app: app})
		}
	}
	return keys, nil
}

func uniqueName(used map[string]bool, name string) string {
	ext := path.Ext(name)
	base := strings.TrimSuffix(name, ext)
	n := name
	for i := 2; used[n]; i++ {
		n = fmt.Sprintf("%s (%d)%s", base, i, ext)
	}
	used[n] = true
	return n
}

// ZIP のファイルを書き出す（既にあれば何もしない）
func writeFile(f *zip.File, dst string) error {

	if _, err := os.Stat(dst); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	r, err := f.Open()
	if err != nil {
		return err
	}
	defer r.Close()

	tmp := dst + ".tmp"
	w, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, r); err != nil {
		w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

var webpName = regexp.MustCompile(`^(\d+)_\d+x\d+\.webp$`)
var sourceName = regexp.MustCompile(`^(\d+)\.(jpg|jpeg|png|webp)$`)

func install(args []string) error {

	fs := flag.NewFlagSet("install", flag.ExitOnError)
	work := fs.String("work", "work", "作業ディレクトリ")
	dest := fs.String("dest", "public/images", "画像の配置先")
	fs.Parse(args)

	src := filepath.Join(*work, "images")
	dirs, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		key := d.Name()
		entries, err := os.ReadDir(filepath.Join(src, key))
		if err != nil {
			return err
		}

		// 番号ごとに webp を優先し、無ければ元の画像を使う
		chosen := make(map[string]string)
		for _, e := range entries {
			if m := webpName.FindStringSubmatch(e.Name()); m != nil {
				chosen[m[1]] = e.Name()
			}
		}
		for _, e := range entries {
			m := sourceName.FindStringSubmatch(e.Name())
			if m == nil {
				continue
			}
			if _, ok := chosen[m[1]]; !ok {
				chosen[m[1]] = e.Name()
				fmt.Fprintf(os.Stderr, "%s/%s: webp が無いので元の画像を使う\n", key, e.Name())
			}
		}
		for _, e := range entries {
			n := strings.TrimSuffix(e.Name(), path.Ext(e.Name()))
			if _, ok := chosen[n]; !ok && !strings.Contains(n, "_") && !strings.HasSuffix(e.Name(), ".tmp") {
				fmt.Fprintf(os.Stderr, "%s/%s: 使える画像が無い\n", key, e.Name())
			}
		}

		for n, name := range chosen {
			ext := path.Ext(name)
			if webpName.MatchString(name) {
				ext = ".webp"
			}
			if err := copyFile(filepath.Join(src, key, name), filepath.Join(*dest, key, n+ext)); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
