package src

import (
	"datastore"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"log"
	"src/logic"
)

type Public struct {}

func (p Public) manageTopHandler(w http.ResponseWriter, r *http.Request) {
	p.pagingTop(w,r,true)
}

func (p Public) topHandler(w http.ResponseWriter, r *http.Request) {
	p.pagingTop(w,r,false)
}

func (p Public) pagingTop(w http.ResponseWriter, r *http.Request,flag bool) {
	site := datastore.GetSite(r)
	if site.Root == "" {
		p.errorPage(w,"Not Found","Root page not found",404)
		return
	}
	logic.GenerateHTML(w,r,site.Root,flag)
}


func (p Public) managePageHandler(w http.ResponseWriter, r *http.Request) {
	p.paging(w,r,true)
}

func (p Public) pageHandler(w http.ResponseWriter, r *http.Request) {
	//TODO HTMLアクセス
	p.paging(w,r,false)
}

func (p Public) paging(w http.ResponseWriter, r *http.Request,flag bool) {
	vars := mux.Vars(r)
	id := vars["key"]
	err := logic.GenerateHTML(w,r,id,flag)
	if err != nil {
		p.errorPage(w,"ERROR:Generate HTML",err.Error(),500)
		return
	}
}

func (p Public) fileHandler(w http.ResponseWriter, r *http.Request) {

	//ファイルを検索
	vars := mux.Vars(r)
	id := vars["key"]

	//表示
	fileData, err := datastore.SelectFileData(r, id)
	if err != nil {
		p.errorPage(w ,"Datastore:FileData Search Error", err.Error(), 500)
		return
	}

	if fileData == nil {
		p.errorPage(w, "Datastore:Not Found FileData Error",id , 404)
		return
	}

	w.Header().Set("Content-Type", fileData.Mime)
	_, err = w.Write(fileData.Content)
	if err != nil {
		p.errorPage(w, "Writing FileData Error", err.Error(), 500)
		return
	}
	return
}

func (p Public) errorPage(w http.ResponseWriter, t string, msg string, num int) {

	dto := struct {
		Title   string
		Message string
		No      int
	}{t, msg, num}

	tmpl, err := template.ParseFiles("templates/error.tmpl")
	if err != nil {
		log.Println("Error Page Parse Error")
		log.Println(err)
		return
	}
	err = tmpl.Execute(w, dto)
	if err != nil {
		log.Println("Error Page Execute Error")
		log.Println(err)
		return
	}
}
