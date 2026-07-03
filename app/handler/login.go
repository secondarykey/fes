package handler

import (
	"app/datastore"
	"app/handler/internal"
	"app/handler/manage"
	"os"
	"strings"

	"fmt"
	"net/http"

	"golang.org/x/xerrors"
	"google.golang.org/api/idtoken"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {

	err := manage.SetSession(w, r, nil)
	if err != nil {
		//TODO エラー
	}

	if redirect := r.URL.Query().Get("redirect"); redirect != "" {
		manage.SetRedirectCookie(w, redirect)
	}

	err = internal.View(w, nil, "authentication.tmpl")
	if err != nil {
		errorPage(w, r, "描画エラー", fmt.Errorf("認証ページの表示に失敗 %v", err), 500)
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
	dao := datastore.NewDao()
	defer dao.Close()

	site, err := dao.SelectSite(ctx, -1)
	if err != nil {
		if err != datastore.SiteNotFoundError {
			errorPage(w, r, "サイト取得エラー", err, 500)
			return
		}
	}

	r.ParseForm()

	tokenString := r.FormValue("credential")

	// Google ID トークンを検証する（署名・有効期限・audience）。
	// audience には自身の CLIENT_ID を指定し、他アプリ向けに発行された
	// トークンでのログインを防ぐ。
	payload, err := idtoken.Validate(ctx, tokenString, os.Getenv("CLIENT_ID"))
	if err != nil {
		errorPage(w, r, "認証エラー", xerrors.Errorf("idtoken.Validate() error: %w", err), 401)
		return
	}

	// 発行者が Google であることを確認
	if payload.Issuer != "accounts.google.com" && payload.Issuer != "https://accounts.google.com" {
		errorPage(w, r, "認証エラー", fmt.Errorf("invalid issuer: %s", payload.Issuer), 401)
		return
	}

	email, _ := payload.Claims["email"].(string)
	if email == "" {
		errorPage(w, r, "認証エラー", fmt.Errorf("email claim not found"), 401)
		return
	}

	if verified, _ := payload.Claims["email_verified"].(bool); !verified {
		errorPage(w, r, "認証エラー", fmt.Errorf("email not verified: %s", email), 403)
		return
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
		//TODO Managers 未設定時は初期セットアップのため全許可（要見直し）
		flag = true
	}

	if !flag {
		errorPage(w, r, "認証エラー", fmt.Errorf("管理者として登録されていません: %s", email), 403)
		return
	}
	//Cookieの作成
	u := manage.NewLoginUser(email, tokenString)
	err = manage.SetSession(w, r, u)
	if err != nil {
		errorPage(w, r, "セッション作成エラー", err, 500)
		return
	}

	redirect := manage.PopRedirectCookie(w, r)
	if redirect == "" || !strings.HasPrefix(redirect, "/manage") {
		redirect = "/manage/"
	}
	http.Redirect(w, r, redirect, 302)
}
