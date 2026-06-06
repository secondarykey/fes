package logic

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"

	"github.com/nfnt/resize"
)

func ConvertImage(r io.Reader) ([]byte, bool, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, false, err
	}

	var img image.Image
	buff := bytes.NewBuffer(b)
	cnv := false
	//over 1mb
	if len(b) > (1 * 1024 * 1024) {
		if img == nil {
			img, _, err = image.Decode(buff)
			if err != nil {
				return nil, false, err
			}
		}

		img = resize.Resize(1000, 0, img, resize.Lanczos3)
		cnv = true
	}

	if cnv {
		buffer := new(bytes.Buffer)
		if err := jpeg.Encode(buffer, img, nil); err != nil {
			return nil, cnv, err
		}
		b = buffer.Bytes()
	}

	return b, cnv, nil
}
