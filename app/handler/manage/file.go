package manage

import (
	"app/datastore"
	"app/handler/manage/form"

	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"golang.org/x/xerrors"
)

func FileViewHandler(w http.ResponseWriter, r *http.Request) error {
	//ファイルを検索
	vars := mux.Vars(r)
	id := vars["key"]

	dao := datastore.NewDao()
	defer dao.Close()

	//表示
	fileData, err := dao.GetFileData(r.Context(), id)
	if err != nil {
		return xerrors.Errorf("GetFileData() error: %w", err)
	}

	if fileData == nil {
		return fmt.Errorf("FileData is nil: %s", id)
	}

	w.Header().Set("Content-Type", fileData.Mime)
	_, err = w.Write(fileData.Content)
	if err != nil {
		return xerrors.Errorf("Writer Write() error: %w", err)
	}
	return nil
}

func fileViewHandler(w http.ResponseWriter, r *http.Request) {
	err := FileViewHandler(w, r)
	if err != nil {
		errorPage(w, "Error file View", err, 404)
	}
}

func viewFileHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	t, flag := vars["type"]
	if !flag {
		t = "1"
	}

	q := r.URL.Query()
	cursor := q.Get("cursor")

	ctx := r.Context()

	dao := datastore.NewDao()
	defer dao.Close()

	files, next, err := dao.SelectFiles(ctx, t, cursor)
	if err != nil {
		errorPage(w, "Error Select File", err, 500)
		return
	}

	dto := struct {
		Files []datastore.File
		Now   string
		Next  string
	}{files, cursor, next}
	viewManage(w, "file/view.tmpl", dto)
}

//URL = /manage/file/add
func addFileHandler(w http.ResponseWriter, r *http.Request) {

	fs := new(datastore.FileSet)

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	fs.File = &datastore.File{}
	fs.FileData = &datastore.FileData{}

	err := form.SetFile(r, fs, datastore.FileTypeData)
	if err != nil {
		errorPage(w, "SetFile() Error", err, 500)
		return
	}

	err = dao.SaveFile(ctx, fs)
	if err != nil {
		errorPage(w, "Error Add File", err, 500)
		return
	}
	//リダイレクト
	http.Redirect(w, r, "/manage/file/", 302)
}

func faviconUploadHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	fs := new(datastore.FileSet)

	err := form.SetFile(r, fs, datastore.FileTypeSystem)
	if err != nil {
		errorPage(w, "SetFile() Error", err, 500)
		return
	}

	fs.ID = datastore.SystemFaviconID
	fs.Name = datastore.SystemFaviconID

	err = dao.SaveFile(ctx, fs)
	if err != nil {
		errorPage(w, "Error Add File", err, 500)
		return
	}
	//リダイレクト
	http.Redirect(w, r, "/manage/site/", 302)
}

//URL = /manage/file/delete
func deleteFileHandler(w http.ResponseWriter, r *http.Request) {

	//リダイレクト
	id := r.FormValue("fileName")
	ctx := r.Context()

	dao := datastore.NewDao()
	defer dao.Close()

	err := dao.RemoveFile(ctx, id)
	if err != nil {
		errorPage(w, "RemoveFile Error", err, 500)
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

func resizeFileHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]

	ctx := r.Context()

	dao := datastore.NewDao()
	defer dao.Close()

	file, err := dao.SelectFile(ctx, id)
	if err != nil {
		errorPage(w, "Select File Error", err, 500)
		return
	}

	dto := struct {
		File *datastore.File
	}{file}
	viewManage(w, "file/resize.tmpl", dto)
}

func resizeCommitFileHandler(w http.ResponseWriter, r *http.Request) {

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

	err := writeResize(writer, r, resize)
	if err != nil {
		errorPage(w, "Resize Error", err, 500)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	err = dao.PutFileData(ctx, resize.id, writer.Bytes(), "image/jpeg")
	if err != nil {
		errorPage(w, "Datastore FileData Put Error", err, 500)
		return
	}

	http.Redirect(w, r, "/manage/file/resize/"+resize.id, 302)
}

func resizeFileViewHandler(w http.ResponseWriter, r *http.Request) {

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

	err := writeResize(w, r, resize)
	if err != nil {
		errorPage(w, "Resize Error", err, 500)
	}
}

func writeResize(w io.Writer, r *http.Request, re Resize) error {

	dao := datastore.NewDao()
	fileData, err := dao.GetFileData(r.Context(), re.id)
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
