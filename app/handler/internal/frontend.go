package internal

// 管理画面（manage/）と会場マップ（maps/）のビルドと、_assets/manage・_assets/maps への反映。
// それぞれのディレクトリで npm run deploy を実行する。
//
//	go generate ./handler/internal
//
//go:generate go run ../../_cmd/frontbuild/main.go
