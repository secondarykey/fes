package app

// Package: このファイルは動作には関係しません
//
// このファイルが存在することでgcloudコマンドでのデプロイを実現しています。
// GOPATHを自身のディレクトリに設定し、シンボリックリンクでライブラリを指すことで
// 可能ですが、app.yamlをsrcに置きたくないので、この構成にしてこのファイルで
// 読み込ませることで実現しています。
//
// References:
//   https://github.com/secondarykey/fes/wiki
import (
	_ "src/api"
	_ "src/datastore"
	_ "src/manage"

	_ "github.com/gorilla/mux"
	_ "github.com/knightso/base/gae/ds"
	_ "github.com/nfnt/resize"
	_ "github.com/satori/go.uuid"
)
