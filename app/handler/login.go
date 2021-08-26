package handler

import (
	"app/datastore"
	. "app/handler/internal"
	"app/handler/manage"
	"os"

	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {

	err := manage.SetSession(w, r, nil)
	if err != nil {
		//TODO エラー
	}

	err = View(w, nil, "authentication.tmpl")
	if err != nil {
		errorPage(w, "描画エラー", fmt.Errorf("認証ページの表示に失敗 %v", err), 500)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	err := manage.ClearSession(w, r)
	if err != nil {
		//TODO エラー
	}
	http.Redirect(w, r, "/login", 302)
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	site, err := datastore.SelectSite(ctx, -1)
	if err != nil {
		if err != datastore.SiteNotFoundError {
			errorPage(w, "サイト取得エラー", err, 500)
			return
		}
	}

	r.ParseForm()

	tokenString := r.FormValue("credential")

	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		secret := os.Getenv("CLIENT_SECRET")
		return secret, nil
	})

	if err != nil {
		// TODO key is of invalid type
		//errorPage(w, "JWT解析エラー", err, 500)
		//return
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

	if !flag {
		errorPage(w, "認証エラー", err, 403)
		return
	} else {
		//Cookieの作成
		u := manage.NewLoginUser(email, tokenString)
		err = manage.SetSession(w, r, u)
		if err != nil {
			errorPage(w, "セッション作成エラー", err, 500)
			return
		}
	}
	http.Redirect(w, r, "/manage/", 302)
}
