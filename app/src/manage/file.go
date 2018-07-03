package manage

import (
	"net/http"
	"datastore"
	"api"

	"strconv"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"image/jpeg"
	"image"
	"log"
	"bytes"
)

//URL = /manage/file/
func (h Handler) ViewFile(w http.ResponseWriter, r *http.Request) {

	p:= 1
	q := r.URL.Query()
	pageBuf := q.Get("page")
	if pageBuf != "" {
		page,err := strconv.Atoi(pageBuf)
		if err == nil {
			p = page
		}
	}

	files,err := datastore.SelectFiles(r,p)
	if err != nil {
		h.errorPage(w,"Error Select File",err.Error(),500)
		return
	}

	dto := struct {
		Files []datastore.File
		Page int
		Prev int
		Next int
	} {files,p,p-1,p+1}
	h.parse(w, TEMPLATE_DIR + "file/view.tmpl", dto)
}

//URL = /manage/file/add
func (h Handler) AddFile(w http.ResponseWriter, r *http.Request) {

	err := datastore.SaveFile(r,"",api.DATA_FILE)
	if err != nil {
		h.errorPage(w,"Error Add File",err.Error(),500)
		return
	}
	//リダイレクト
	http.Redirect(w, r, "/manage/file/", 302)
}

//URL = /manage/file/delete
func (h Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	//リダイレクト
	id := r.FormValue("fileName")
	err := datastore.RemoveFile(r,id)
	if err != nil {
		h.errorPage(w,err.Error(),id,500)
		return
	}
	http.Redirect(w, r, "/manage/file/", 302)
}

func (h Handler) ResizeFile(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]
	file,err := datastore.SelectFile(r,id)
	if err != nil {
		h.errorPage(w,err.Error(),id,500)
		return
	}

	dto := struct {
		File *datastore.File
	} {file}
	h.parse(w, TEMPLATE_DIR + "file/resize.tmpl", dto)
}

func (h Handler) ResizeFileView(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]

	fileData, err := datastore.SelectFileData(r, id)
	if err != nil {
		h.errorPage(w, "Datastore:FileData Search Error", err.Error(), 500)
		return
	}

	if fileData == nil {
		h.errorPage(w, "Datastore:Not Found FileData Error", id, 404)
		return
	}

	q := r.URL.Query()

	leftBuf := q.Get("left")
	topBuf := q.Get("top")
	widthBuf := q.Get("width")
	heightBuf := q.Get("height")

	var img image.Image

	buff := bytes.NewBuffer(fileData.Content)
	//元データのポインタを作成
	img, _, err = image.Decode(buff)
	if err != nil {
		return
	}

	//すべてが０だった場合、やらなくていい
	if !zero(widthBuf) && !zero(heightBuf) {

		//新しいサイズを作成
		left,_ := strconv.ParseInt(leftBuf,10,64)
		top,_ := strconv.ParseInt(topBuf,10,64)
		width,_ := strconv.ParseInt(widthBuf,10,64)
		height,_ := strconv.ParseInt(heightBuf,10,64)

		img = cut(img,int(left),int(top),int(width),int(height))
	}
	//現在の画像から幅を取得
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	perBuf := q.Get("per")
	per, err := strconv.ParseInt(perBuf, 10, 64)
	if err != nil {
		return
	}

	//方法を変換
	function := getJPEGFunction(q.Get("function"))
	//TODO perが100の場合やらなくていいってわけでもないかな？
	newWidth :=  float64(width) * float64(per) / 100
	newHeight := float64(height) * float64(per) / 100
	newImg := resize.Resize(uint(newWidth), uint(newHeight), img, function)

	quaBuf := q.Get("quality")
	qua,err := strconv.ParseInt(quaBuf,10,64)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", fileData.Mime)
	err = jpeg.Encode(w,newImg,&jpeg.Options{Quality:int(qua)})
	if err != nil {
		return
	}
}

func cut(org image.Image,l,t,w,h int) image.Image {
	r := image.Rect(0,0,w-l,h-t)
	dst := image.NewNRGBA(r)
	x := 0
	for dx := l; dx <= w; dx++ {
		y := 0
		for dy := t; dy <= h; dy++ {
			dst.Set(x,y,org.At(dx,dy))
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
	_ ,err := strconv.ParseInt(b,10,64)
	if err != nil {
		return true
	}
	return false
}

func getJPEGFunction(b string) resize.InterpolationFunction {
	log.Println(b)
	switch b {
	case "Bilinear" : return resize.Bilinear
	case "Bicubic" : return resize.Bicubic
	case "Lanczos2" : return resize.Lanczos2
	case "Lanczos3" : return resize.Lanczos3
	case "MitchellNetravali" : return resize.MitchellNetravali
	case "NearestNeighbor" : return resize.NearestNeighbor
	}
	return resize.Lanczos3
}
