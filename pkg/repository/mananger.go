package repository

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog"
	"github.com/valerius21/gitignore.lol/pkg/utils"
)

type Repository struct {
	logger *zerolog.Logger
}

func (r *Repository) InitRepoWatch() (*gocron.Scheduler, error) {
	// init logging
	logger := utils.InitLogger()
	r.logger = &logger

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	job, err := scheduler.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func(a string) {
				r.logger.Info().Msg(a)
			},
			"hello",
		),
	)
	if err != nil {
		return nil, err
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine to handle shutdown
	go func() {
		<-sigChan
		r.logger.Info().Msg("Shutting down scheduler...")
		scheduler.Shutdown()
		r.logger.Info().Msg("Scheduler shut down")
		close(sigChan)
	}()

	r.logger.Info().Msgf("Job ID: %v", job.ID())

	return &scheduler, nil
}

func (r *Repository) Close(s *gocron.Scheduler) {
	r.logger.Info().Msgf("Shutting down scheduler")
	(*s).Shutdown()
}
