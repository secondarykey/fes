package manage

import (
	. "app/handler/internal"

	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
)

func registerV2(s *mux.Router) error {
	if err := RegisterManageV2Static(); err != nil {
		return xerrors.Errorf("RegisterManageV2Static() error: %w", err)
	}

	// API ルートは SPA キャッチオールより先に登録する
	api := s.PathPrefix("/v2/api").Subrouter()
	registerV2PageAPI(api)
	registerV2FileAPI(api)
	registerV2TemplateAPI(api)
	registerV2SiteAPI(api)
	registerV2VariableAPI(api)
	registerV2DraftAPI(api)
	registerV2ToolAPI(api)

	// SPA キャッチオール — /v2/ 配下のすべての非API/非アセットパスに index.html を返す
	s.PathPrefix("/v2/").HandlerFunc(ServeManageV2SPA)

	return nil
}
