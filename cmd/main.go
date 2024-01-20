package main

import (
	"fmt"
	"os"

	"github.com/valerius21/gitignore.lol/pkg/web"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	ws := web.DefaultWebServer

	ws.App.Listen(fmt.Sprintf(":%s", port))
}
