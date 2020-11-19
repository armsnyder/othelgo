package main

import (
	"flag"
	"log"

	"github.com/armsnyder/othelgo/pkg/client"
)

func main() {
	local := flag.Bool("local", false, "If true, connect to a local server.")
	flag.Parse()

	if err := client.Run(*local); err != nil {
		log.Fatal(err)
	}
}
