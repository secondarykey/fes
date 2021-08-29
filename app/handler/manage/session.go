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

func getSessionOptions(age int) *sessions.Options {
	return &sessions.Options{
		Path:     "/",
		MaxAge:   age,
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
