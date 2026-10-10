package app

// 管理画面（manage/）と会場マップ（maps/）のビルドと、app/handler/internal/_assets/manage・_assets/maps への反映。
// それぞれのディレクトリで npm run deploy を実行する（app/_cmd/frontbuild）。
//
//	cd app
//	go generate ./...
//
//go:generate go run ./_cmd/frontbuild
