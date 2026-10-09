// frontbuild は管理画面（manage/）と会場マップ（maps/）をビルドし、
// 成果物を Go の embed 対象ディレクトリ（app/handler/internal/_assets/<名前>/）へ反映するツール。
//
// 各ディレクトリで npm run deploy（lint → ビルド → embed 用ディレクトリへのコピー）を実行する。
// node_modules が無ければ先に npm install する。
//
// app/handler/internal/frontend.go の go:generate から呼ばれる想定で、
// デフォルトのパスは同ディレクトリ（app/handler/internal）を基準にしている。
//
//	go generate ./handler/internal
//	go run ./_cmd/frontbuild -root .. -only maps    （app/ から maps だけを反映する例）
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// 反映する SPA（リポジトリ直下のディレクトリ名）
var targets = []string{"manage", "maps"}

func main() {
	root := flag.String("root", filepath.Join("..", "..", ".."), "リポジトリのルート")
	only := flag.String("only", "", "反映する SPA をカンマ区切りで指定する（manage,maps）。空なら全部")
	flag.Parse()

	names := targets
	if *only != "" {
		names = strings.Split(*only, ",")
	}

	for _, name := range names {
		if err := deploy(*root, strings.TrimSpace(name)); err != nil {
			log.Fatalf("frontbuild: %+v", err)
		}
	}
}

func deploy(root, name string) error {

	dir, err := filepath.Abs(filepath.Join(root, name))
	if err != nil {
		return err
	}
	if stat, err := os.Stat(filepath.Join(dir, "package.json")); err != nil || stat.IsDir() {
		return fmt.Errorf("%s の package.json が見つかりません: %s", name, dir)
	}

	if _, err := os.Stat(filepath.Join(dir, "node_modules")); os.IsNotExist(err) {
		if err := npm(dir, "install"); err != nil {
			return err
		}
	}
	return npm(dir, "run", "deploy")
}

func npm(dir string, args ...string) error {

	// Windows では npm.cmd が解決されるため LookPath 経由で実行する
	bin, err := exec.LookPath("npm")
	if err != nil {
		return fmt.Errorf("npm が見つかりません: %w", err)
	}

	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("frontbuild: npm %s (%s)\n", strings.Join(args, " "), dir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm %s (%s): %w", strings.Join(args, " "), dir, err)
	}
	return nil
}
