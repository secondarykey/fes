package main

import (
	"app"
	"app/config"

	"fmt"
	"log"
)

//
// このコマンドはローカルのデータから静的なサイトを出力します
//
// 指定されたディレクトリ名でURLにして切り替えます。
// 対象はFileDataとHTMLになります。
//
func main() {
	err := app.CreateStaticSite(
		"2022-Spring",
		config.SetProjectID(),
		config.SetDatastore())
	if err != nil {
		log.Fatalf("%+v", err)
	}
	fmt.Println("Success!")
}
