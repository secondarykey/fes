// managebuild は管理画面 SPA (manage/) をビルドし、
// 成果物を Go の embed 対象ディレクトリへ同期するツール。
//
// app/handler/internal/manage_spa.go の go:generate から呼ばれる想定で、
// デフォルトのパスは同ディレクトリ（app/handler/internal）を基準にしている。
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	src := flag.String("src", filepath.Join("..", "..", "..", "manage"), "管理画面 SPA のソースディレクトリ")
	dst := flag.String("dst", filepath.Join("_assets", "manage"), "embed 対象の出力先ディレクトリ")
	skipBuild := flag.Bool("skip-build", false, "npm run build を行わず既存の dist だけを同期する")
	flag.Parse()

	if err := run(*src, *dst, *skipBuild); err != nil {
		log.Fatalf("managebuild: %+v", err)
	}
}

func run(src, dst string, skipBuild bool) error {

	srcDir, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	dstDir, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	if stat, err := os.Stat(srcDir); err != nil || !stat.IsDir() {
		return fmt.Errorf("ソースディレクトリが見つかりません: %s", srcDir)
	}

	if !skipBuild {
		if _, err := os.Stat(filepath.Join(srcDir, "node_modules")); os.IsNotExist(err) {
			if err := npm(srcDir, "install"); err != nil {
				return err
			}
		}
		if err := npm(srcDir, "run", "build"); err != nil {
			return err
		}
	}

	distDir := filepath.Join(srcDir, "dist")
	if stat, err := os.Stat(distDir); err != nil || !stat.IsDir() {
		return fmt.Errorf("ビルド成果物が見つかりません: %s", distDir)
	}

	// 旧ファイルが残ると embed に混ざるため、出力先は作り直す
	if err := os.RemoveAll(dstDir); err != nil {
		return err
	}
	if err := copyDir(distDir, dstDir); err != nil {
		return err
	}

	fmt.Printf("managebuild: %s -> %s\n", distDir, dstDir)
	return nil
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

	fmt.Printf("managebuild: npm %v (%s)\n", args, dir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm %v: %w", args, err)
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
