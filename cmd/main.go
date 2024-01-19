package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/oklog/run"
	"github.com/valerius21/gitignore.lol/pkg/repository"
	"github.com/valerius21/gitignore.lol/pkg/utils"
	"github.com/valerius21/gitignore.lol/pkg/web"
)

func main() {
	// init logging
	logger := utils.InitLogger()
	var g run.Group
	repo := new(repository.Repository)

	// start repo watch
	sRef, err := repo.InitRepoWatch()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Add scheuler to run group
	{
		scheuler := *sRef
		g.Add(func() error {
			scheuler.Start()
			<-ctx.Done()
			return ctx.Err()
		}, func(error) {
			scheuler.Shutdown()
			cancel()
		})
	}

	// Start webserver
	server := new(web.WebServer)
	{
		g.Add(func() error {
			server.Init()
			return server.App.Listen(":3000")
		}, func(error) {
			if err := server.App.Shutdown(); err != nil {
				logger.Error().Msgf("Error shutting down Fiber app: %v", err)
			}
		})
	}

	// Handle OS signals
	{
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		g.Add(func() error {
			select {
			case sig := <-sigChan:
				logger.Info().Msgf("Received signal %s", sig)
				return nil
			}
		}, func(error) {
			close(sigChan)
		})
	}

	// Start run group
	if err := g.Run(); err != nil {
		logger.Error().Msgf("Error running run group: %v", err)
	}
}
