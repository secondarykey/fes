package main

import (
	"app"
	"app/config"

	"fmt"
	"log"
)

func main() {
	err := app.CreateStaticSite(
		"2020",
		config.SetProjectID(),
		config.SetDatastore())
	if err != nil {
		log.Fatalf("%+v", err)
	}
	fmt.Println("Success!")
}
