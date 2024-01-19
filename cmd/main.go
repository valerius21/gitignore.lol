package main

import (
	"github.com/valerius21/gitignore.lol/pkg/repository"
	"github.com/valerius21/gitignore.lol/pkg/web"
)

func main() {
	// init logging
	_ = repository.DefaultRepository
	ws := web.DefaultWebServer

	ws.App.Listen(":3000")
}
