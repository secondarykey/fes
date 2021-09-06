package main

import (
	"app"
	"app/config"

	"fmt"
	"log"
	"os"
)

func main() {
	err := app.Listen(
		config.SetProjectID(),
		config.SetDatastore())
	if err != nil {
		log.Fatalf("%+v", err)
		os.Exit(1)
	}
	fmt.Println("bye!")
	return
}
