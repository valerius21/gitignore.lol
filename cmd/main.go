// Package main is the entry point for the gitignore.lol application
package main

import (
	"fmt"

	"github.com/alecthomas/kong"

	"me.valerius/gitignore-lol/pkg/lib"
	"me.valerius/gitignore-lol/pkg/server"
)

// @Summary      Main entry point
// @Description  Initializes the gitignore.lol application
func main() {
	// Parse command line flags
	_ = kong.Parse(&lib.CLI)

	gr := lib.NewGitRunner(lib.CLI.BaseRepository, lib.CLI.ClonePath, lib.CLI.UpdateInterval)
	lib.Logger.Info("CLI ARGS", "repository", lib.CLI.BaseRepository)

	err := gr.Init()
	if err != nil {
		lib.Logger.Error("Failed to initialize Git Repository", "error", err)
		panic(1)
	}

	lib.Logger.Info(fmt.Sprintf("Gitignores cloned to %s\n", gr.LocalPath))

	// Start the server with the configured port
	if err := server.Run(lib.CLI.Port, gr); err != nil {
		lib.Logger.Error("Server failed to start", "error", err)
	}
}
