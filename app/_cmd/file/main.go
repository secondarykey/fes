package main

import (
	"app"
	"app/config"

	"fmt"
	"os"

	"golang.org/x/xerrors"
)

func main() {
	err := run()
	if err != nil {
		fmt.Printf("run error:\n%+v\n", err)
		os.Exit(1)
	}
	fmt.Println("Success")
}

func run() error {

	dir := "20210608"
	//datastoreの準備
	err := app.GenerateFiles(
		dir,
		config.SetProjectID(),
		config.SetDatastore())
	if err != nil {
		return xerrors.Errorf("app.GenerateFiles() error: %w", err)
	}
	return nil
}
