package main

import (
	"github.com/valerius21/gitignore.lol/pkg/web"
)

func main() {
	ws := web.DefaultWebServer

	ws.App.Listen(":3000")
}
