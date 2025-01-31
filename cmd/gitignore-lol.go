package main

import (
	"log"

	"github.com/alecthomas/kong"
	"me.valerius/gitignore-lol/pkg/lib"
	"me.valerius/gitignore-lol/pkg/server"
)

func main() {
	// Parse command line flags
	_ = kong.Parse(&lib.CLI)

	// Start the server with the configured port
	if err := server.Run(lib.CLI.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
