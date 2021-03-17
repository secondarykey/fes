package manage

import (
	"app/api"
	"app/datastore"
	"fmt"

	"bytes"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
)

//URL = /manage/file/
func (h Handler) ViewFile(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	t, flag := vars["type"]
	if !flag {
		t = "1"
	}

	p := 1
	q := r.URL.Query()
	pageBuf := q.Get("page")
	if pageBuf != "" {
		page, err := strconv.Atoi(pageBuf)
		if err == nil {
			p = page
		}
	}

	files, err := datastore.SelectFiles(r, t, p)
	if err != nil {
		h.errorPage(w, "Error Select File", err, 500)
		return
	}

	dto := struct {
		Files []datastore.File
		Page  int
		Prev  int
		Next  int
	}{files, p, p - 1, p + 1}
	h.parse(w, TEMPLATE_DIR+"file/view.tmpl", dto)
}

//URL = /manage/file/add
func (h Handler) AddFile(w http.ResponseWriter, r *http.Request) {

	err := datastore.SaveFile(r, "", api.FileTypeData)
	if err != nil {
		h.errorPage(w, "Error Add File", err, 500)
		return
	}
	//リダイレクト
	http.Redirect(w, r, "/manage/file/", 302)
}

//URL = /manage/file/delete
func (h Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	//リダイレクト
	id := r.FormValue("fileName")
	err := datastore.RemoveFile(r, id)
	if err != nil {
		h.errorPage(w, "RemoveFile Error", err, 500)
		return
	}
	http.Redirect(w, r, "/manage/file/", 302)
}

type Resize struct {
	id       string
	left     string
	top      string
	width    string
	height   string
	per      string
	function string
	quality  string
}

func (h Handler) ResizeFile(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]
	file, err := datastore.SelectFile(r, id)
	if err != nil {
		h.errorPage(w, "Select File Error", err, 500)
		return
	}

	dto := struct {
		File *datastore.File
	}{file}
	h.parse(w, TEMPLATE_DIR+"file/resize.tmpl", dto)
}

func (h Handler) ResizeCommitFile(w http.ResponseWriter, r *http.Request) {

	resize := Resize{
		id:       r.FormValue("key"),
		left:     r.FormValue("left"),
		top:      r.FormValue("top"),
		width:    r.FormValue("width"),
		height:   r.FormValue("height"),
		per:      r.FormValue("per"),
		function: r.FormValue("function"),
		quality:  r.FormValue("quality"),
	}

	writer := bytes.NewBuffer([]byte(""))
	err := h.WriteResize(writer, r, resize)
	if err != nil {
		h.errorPage(w, "Resize Error", err, 500)
	}

	err = datastore.PutFileData(r, resize.id, writer.Bytes(), "image/jpeg")
	if err != nil {
		h.errorPage(w, "Datastore FileData Put Error", err, 500)
	}

	http.Redirect(w, r, "/manage/file/resize/"+resize.id, 302)
}

func (h Handler) ResizeFileView(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	q := r.URL.Query()

	resize := Resize{
		id:       vars["key"],
		left:     q.Get("left"),
		top:      q.Get("top"),
		width:    q.Get("width"),
		height:   q.Get("height"),
		per:      q.Get("per"),
		function: q.Get("function"),
		quality:  q.Get("quality"),
	}

	err := h.WriteResize(w, r, resize)
	if err != nil {
		h.errorPage(w, "Resize Error", err, 500)
	}
}

func (h Handler) WriteResize(w io.Writer, r *http.Request, re Resize) error {

	fileData, err := datastore.GetFileData(r.Context(), re.id)
	if err != nil {
		return err
	}

	if fileData == nil {
		return err
	}

	var img image.Image

	buff := bytes.NewBuffer(fileData.Content)
	//元データのポインタを作成
	img, _, err = image.Decode(buff)
	if err != nil {
		return err
	}

	//すべてが０だった場合、やらなくていい
	if !zero(re.width) && !zero(re.height) {

		//新しいサイズを作成
		left, _ := strconv.ParseInt(re.left, 10, 64)
		top, _ := strconv.ParseInt(re.top, 10, 64)
		width, _ := strconv.ParseInt(re.width, 10, 64)
		height, _ := strconv.ParseInt(re.height, 10, 64)

		img = cut(img, int(left), int(top), int(width), int(height))
	}
	//現在の画像から幅を取得
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	per, err := strconv.ParseInt(re.per, 10, 64)
	if err != nil {
		return err
	}

	//方法を変換
	function := getJPEGFunction(re.function)
	//TODO perが100の場合やらなくていいってわけでもないかな？
	newWidth := float64(width) * float64(per) / 100
	newHeight := float64(height) * float64(per) / 100
	newImg := resize.Resize(uint(newWidth), uint(newHeight), img, function)

	qua, err := strconv.ParseInt(re.quality, 10, 64)
	if err != nil {
		return err
	}

	if res, ok := w.(http.ResponseWriter); ok {
		res.Header().Set("Content-Type", fileData.Mime)
	}
	err = jpeg.Encode(w, newImg, &jpeg.Options{Quality: int(qua)})
	if err != nil {
		return err
	}
	return nil
}

func cut(org image.Image, l, t, w, h int) image.Image {
	r := image.Rect(0, 0, w-l, h-t)
	dst := image.NewNRGBA(r)
	x := 0
	for dx := l; dx <= w; dx++ {
		y := 0
		for dy := t; dy <= h; dy++ {
			dst.Set(x, y, org.At(dx, dy))
			y++
		}
		x++
	}
	return dst
}

func zero(b string) bool {
	if b == "" || b == "0" {
		return true
	}
	_, err := strconv.ParseInt(b, 10, 64)
	if err != nil {
		return true
	}
	return false
}

func getJPEGFunction(b string) resize.InterpolationFunction {
	fmt.Println(b)
	switch b {
	case "Bilinear":
		return resize.Bilinear
	case "Bicubic":
		return resize.Bicubic
	case "Lanczos2":
		return resize.Lanczos2
	case "Lanczos3":
		return resize.Lanczos3
	case "MitchellNetravali":
		return resize.MitchellNetravali
	case "NearestNeighbor":
		return resize.NearestNeighbor
	}
	return resize.Lanczos3
}
