package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"me.valerius/gitignore-lol/pkg/lib"
	"me.valerius/gitignore-lol/pkg/server"
)

func main() {
	// TODO: read cli flags, port, base-repo, etc
	ctx := kong.Parse(&lib.CLI)
	fmt.Println(ctx.Command())
	// TODO: startup function
	server.Run(3000)
}
