package handler

import (
	"app/datastore"
	. "app/handler/internal"
	"app/handler/manage"
	"fmt"

	"encoding/json"
	"log"
	"net/http"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {

	err := manage.SetSession(w, r, nil)
	if err != nil {

	}

	err = View(w, nil, "authentication.tmpl")
	if err != nil {
		errorPage(w, "描画エラー", fmt.Errorf("認証ページの表示に失敗"), 500)
	}
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {

	code := 200
	dto := struct {
		Success bool
	}{false}

	ctx := r.Context()
	site, err := datastore.SelectSite(ctx, -1)
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

		err := manage.SetSession(w, r, u)
		if err != nil {
			code = 500
			dto.Success = false
			log.Println(err)
		}
	}

	w.WriteHeader(code)

	err = json.NewEncoder(w).Encode(dto)
	if err != nil {
		log.Println(err)
	}
}
