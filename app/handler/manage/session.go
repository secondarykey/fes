package manage

import (
	"app/config"

	"crypto/rand"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"golang.org/x/xerrors"
)

var store *sessions.CookieStore

func init() {
	gob.Register(&LoginUser{})
}

// InitSessionStore はセッションストアを初期化する。
// 環境変数 SESSION_KEY を署名鍵に使用する。環境変数のセット後
// （setEnvironment 実行後）に呼ぶ必要があるため、パッケージ init ではなく
// Register() から明示的に呼び出す。
func InitSessionStore() error {
	key := os.Getenv("SESSION_KEY")
	if key == "" {
		if !config.Get().DevelopMode {
			return xerrors.Errorf("SESSION_KEY is not set")
		}
		// 開発用: ランダム鍵を生成する（再起動でセッションは無効化される）
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return xerrors.Errorf("rand.Read() error: %w", err)
		}
		key = string(b)
		log.Println("SESSION_KEY is not set; using an ephemeral random key (development only)")
	}
	store = sessions.NewCookieStore([]byte(key))
	return nil
}

const sessionName = "session"

type LoginUser struct {
	Email string
	Token string
}

func getSessionOptions(age int) *sessions.Options {
	return &sessions.Options{
		Path:     "/",
		MaxAge:   age,
		HttpOnly: true,
		// CSRF 対策: クロスサイトリクエストに Cookie を載せない。
		// Lax はトップレベル GET ナビゲーションのみ許可（ログイン後のリダイレクトに必要）
		SameSite: http.SameSiteLaxMode,
		// ローカル開発時は http のため無効化
		Secure: !config.Get().DevelopMode,
	}
}

func NewLoginUser(email string, token string) *LoginUser {
	user := LoginUser{}
	user.Email = email
	user.Token = token
	return &user
}

func GetSession(r *http.Request) (*LoginUser, error) {
	sess, err := store.Get(r, sessionName)
	if err != nil {
		return nil, xerrors.Errorf("store.Get() error: %w", err)
	}

	obj := sess.Values["User"]
	if user, ok := obj.(*LoginUser); ok {
		return user, nil
	}
	return nil, fmt.Errorf("ユーザの取得失敗")
}

func SetSession(w http.ResponseWriter, r *http.Request, u *LoginUser) error {

	sess, err := store.Get(r, sessionName)
	if err != nil {
		return xerrors.Errorf("store.Get() error: %w", err)
	}

	age := 86400 * 7
	if u == nil {
		age = -1
	}

	sess.Options = getSessionOptions(age)
	sess.Values["User"] = u

	return sess.Save(r, w)
}

func ClearSession(w http.ResponseWriter, r *http.Request) error {
	return SetSession(w, r, nil)
}

func SetDraftId(w http.ResponseWriter, r *http.Request, id string) error {

	sess, err := store.Get(r, sessionName)
	if err != nil {
		return xerrors.Errorf("store.Get() error: %w", err)
	}

	age := 86400 * 7
	sess.Options = getSessionOptions(age)
	sess.Values["DraftId"] = id

	return sess.Save(r, w)
}

func GetDraftId(r *http.Request) (string, error) {
	sess, err := store.Get(r, sessionName)
	if err != nil {
		return "", xerrors.Errorf("store.Get() error: %w", err)
	}

	obj := sess.Values["DraftId"]
	if id, ok := obj.(string); ok {
		return id, nil
	}
	return "", fmt.Errorf("下書きのID取得失敗")
}

func SetRedirectCookie(w http.ResponseWriter, redirect string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "login_redirect",
		Value:    redirect,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	})
}

func PopRedirectCookie(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie("login_redirect")
	if err != nil || c.Value == "" {
		return ""
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "login_redirect",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	return c.Value
}
