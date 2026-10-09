// HEIC の EXIF の向きを読み、向きを直した {名前}_rotate.png を書き出す（向きが 1 なら何もしない）
//
// webp 化（D:\Go\Tools\webp）は HEIC の向きを反映しないため、tools/images extract の後に使う。
// goheif が cgo を使うので gcc が必要。go.mod を持つ別モジュールにしている。
//
//	cd maps/tools/heicrot
//	go run . <work/images/{key}/{n}.heif ...>
//
// EXIF の向きが実際の写真と合っていないこともあるので、書き出したものは目視で確認する。
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/jdeng/goheif"
	"github.com/rwcarlsen/goexif/exif"
)

func main() {
	for _, p := range os.Args[1:] {
		if err := run(p); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", p, err)
		}
	}
}

func run(p string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()

	o := 1
	if raw, err := goheif.ExtractExif(f); err == nil {
		if x, err := exif.Decode(bytes.NewReader(raw)); err == nil {
			if t, err := x.Get(exif.Orientation); err == nil {
				o, _ = t.Int(0)
			}
		}
	}
	fmt.Printf("%s orientation=%d\n", p, o)
	if o == 1 {
		return nil
	}

	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	img, err := goheif.Decode(f)
	if err != nil {
		return err
	}
	dst := apply(img, o)
	out := strings.TrimSuffix(p, filepath.Ext(p)) + "_rotate.png"
	w, err := os.Create(out)
	if err != nil {
		return err
	}
	defer w.Close()
	return png.Encode(w, dst)
}

// EXIF の Orientation（1〜8）に従って向きを直す
func apply(src image.Image, o int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	nw, nh := w, h
	if o >= 5 {
		nw, nh = h, w
	}
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var nx, ny int
			switch o {
			case 2:
				nx, ny = w-1-x, y
			case 3:
				nx, ny = w-1-x, h-1-y
			case 4:
				nx, ny = x, h-1-y
			case 5:
				nx, ny = y, x
			case 6:
				nx, ny = h-1-y, x
			case 7:
				nx, ny = h-1-y, w-1-x
			case 8:
				nx, ny = y, w-1-x
			default:
				nx, ny = x, y
			}
			dst.Set(nx, ny, src.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}
