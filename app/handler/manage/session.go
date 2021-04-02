package manage

import (
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/xerrors"
)

var store = sessions.NewCookieStore([]byte("Let's Festival"))

func init() {
	gob.Register(&LoginUser{})
}

const sessionName = "session"

type LoginUser struct {
	Email string
	Token string
}

func getSessionOptions() *sessions.Options {
	return &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
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

	sess.Options = getSessionOptions()
	sess.Values["User"] = u

	return sess.Save(r, w)
}
