package main

import (
	"app"
	"app/config"
	"fmt"
	"log"
)

func main() {
	err := app.Listen(
		config.SetProjectID(),
		config.SetDatastore())
	if err != nil {
		log.Fatalf("%+v", err)
	}
	fmt.Println("byt!")
	return
}
