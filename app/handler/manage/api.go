package manage

import (
	"app/handler/internal"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
)

func registerAPI(s *mux.Router, root *http.ServeMux) error {
	if err := internal.RegisterManageSPAStatic(root); err != nil {
		return xerrors.Errorf("RegisterManageSPAStatic() error: %w", err)
	}

	// API ルートは SPA キャッチオールより先に登録する
	api := s.PathPrefix("/api").Subrouter()
	registerPageAPI(api)
	registerFileAPI(api)
	registerTemplateAPI(api)
	registerSiteAPI(api)
	registerVariableAPI(api)
	registerDraftAPI(api)
	registerToolAPI(api)

	// SPA キャッチオール — /manage/ 配下のすべての非API/非アセット/非v1パスに index.html を返す
	s.PathPrefix("/").HandlerFunc(internal.ServeManageSPA)

	return nil
}
