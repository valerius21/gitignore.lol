package main

import (
	"github.com/valerius21/gitignore.lol/pkg/repository"
	"github.com/valerius21/gitignore.lol/pkg/utils"
)

func main() {
	// init logging
	logger := utils.InitLogger()
	repo := new(repository.Repository)

	// start repo watch
	sRef, err := repo.InitRepoWatch()
	if err != nil {
		panic(err)
	}

}
