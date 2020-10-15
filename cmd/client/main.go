package main

import (
	"log"

	"github.com/armsnyder/othelgo/pkg/client"
)

func main() {
	if err := client.Run(); err != nil {
		log.Fatal(err)
	}
}
