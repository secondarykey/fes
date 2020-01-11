package main

import (
	"app"
	"log"
)

func main() {
	err := app.Start()
	if err != nil {
		log.Println(err)
	}
}
