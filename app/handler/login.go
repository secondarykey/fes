package handler

import (
	"app/datastore"
	. "app/handler/internal"
	"app/handler/manage"
	"os"

	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {

	err := manage.SetSession(w, r, nil)
	if err != nil {

	}

	err = View(w, nil, "authentication.tmpl")
	if err != nil {
		errorPage(w, "描画エラー", fmt.Errorf("認証ページの表示に失敗 %v", err), 500)
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

	tokenString := r.FormValue("credential")

	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		secret := os.Getenv("CLIENT_SECRET")
		return []byte(secret), nil
	})
	if err != nil {
	}

	emailV, ok := claims["email"]
	email := ""
	if ok {
		email = fmt.Sprintf("%v", emailV)
	}

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
		u := manage.NewLoginUser(email, tokenString)

		err = manage.SetSession(w, r, u)
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
