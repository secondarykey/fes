package handler

import (
	"app/datastore"
	"app/handler/manage"

	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Public struct {
	r *mux.Router
}

func (h Public) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}

func (p Public) loginHandler(w http.ResponseWriter, r *http.Request) {

	err := manage.SetSession(w, r, nil)
	if err != nil {
	}

	tmpl, err := template.ParseFiles("cmd/templates/authentication.tmpl")
	if err != nil {
		log.Println("Error Page Parse Error")
		log.Println(err)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Println("Error Page Execute Error")
		log.Println(err)
		return
	}

}

func (p Public) sessionHandler(w http.ResponseWriter, r *http.Request) {

	code := 200
	dto := struct {
		Success bool
	}{false}

	site, err := datastore.SelectSite(r, -1)
	if err != nil {
		if err != datastore.SiteNotFoundError {
			dto.Success = false
			code = 500
			log.Println(err)
			w.WriteHeader(code)
			json.NewEncoder(w).Encode(dto)
			return
		}
	}

	r.ParseForm()
	email := r.FormValue("email")
	token := r.FormValue("token")

	log.Println(email)
	flag := false

	if site != nil && len(site.Managers) != 0 {
		for _, mail := range site.Managers {
			if email == mail {
				flag = true
				break
			}
		}
	} else {
		flag = true
	}

	dto.Success = flag

	if !flag {
		//403を返す
		code = 403
	} else {
		//Cookieの作成
		u := manage.NewLoginUser(email, token)

		log.Println("write")

		err := manage.SetSession(w, r, u)
		if err != nil {
			code = 500
			dto.Success = false
			log.Println(err)
		}
	}

	w.WriteHeader(code)
	json.NewEncoder(w).Encode(dto)

}

func (p Public) pageHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "public, max-age=3600")
	vars := mux.Vars(r)
	id := vars["key"]
	p.pageView(w, r, id)
}

func (p Public) pageView(w http.ResponseWriter, r *http.Request, id string) {

	w.Header().Set("Cache-Control", "public, max-age=3600")
	//ページを取得してIDを作成
	val := r.URL.Query()
	page := val.Get("page")
	if page != "" {
		id += "?page=" + page
	}

	html, err := datastore.GetHTML(r.Context(), id)
	if err != nil {
		p.errorPage(w, "error get html", err.Error(), 500)
		return
	}
	if html == nil {
		p.errorPage(w, "page not found", id, 404)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	_, err = w.Write(html.Content)
	if err != nil {
		log.Println("Write Error")
	}
}

func (p Public) topHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=3600")
	site, err := datastore.SelectSite(r, -1)
	if err != nil {
		p.errorPage(w, "Not Found", "Root page not found", 404)
		return
	}

	p.pageView(w, r, site.Root)
}

func (p Public) fileHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "public, max-age=3600")
	//ファイルを検索
	vars := mux.Vars(r)
	id := vars["key"]

	//表示
	fileData, err := datastore.GetFileData(r.Context(), id)
	if err != nil {
		p.errorPage(w, "Datastore:FileData Search Error", err.Error(), 500)
		return
	}

	if fileData == nil {
		p.errorPage(w, "Datastore:Not Found FileData Error", id, 404)
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

func (p Public) fileDateCacheHandler(w http.ResponseWriter, r *http.Request) {
	// 60 * 60 * 24 = 86400
	// * 10 = 864000
	w.Header().Set("Cache-Control", "public, max-age=864000")
	p.fileHandler(w, r)
}

func (p Public) fileCacheHandler(w http.ResponseWriter, r *http.Request) {
	// 60 * 60 * 3  = 10800
	// 60 * 60 * 6  = 21600
	// 60 * 60 * 12 = 43200
	// 60 * 60 * 24 = 86400
	w.Header().Set("Cache-Control", "public, max-age=21600")
	p.fileHandler(w, r)
}

func (p Public) sitemap(w http.ResponseWriter, r *http.Request) {
	// 60 * 60 * 24
	w.Header().Set("Cache-Control", "public, max-age=86400")
	err := datastore.GenerateSitemap(w, r)
	if err != nil {
		p.errorPage(w, "Generate sitemap error", err.Error(), 500)
	}
}

func (p Public) errorPage(w http.ResponseWriter, t string, msg string, num int) {
	dto := struct {
		Title   string
		Message string
		No      int
	}{t, msg, num}

	tmpl, err := template.ParseFiles("cmd/templates/error.tmpl")
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
