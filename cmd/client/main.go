package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/armsnyder/othelgo/pkg/client"
)

// version is set at build time using ldflags.
var version = "v0.0.0"

func main() {
	local := flag.Bool("local", false, "If true, connect to a local server.")
	printVersion := flag.Bool("version", false, "Print the client version.")
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		return
	}

	if err := client.Run(*local, version); err != nil {
		log.Fatal(err)
	}
}
