package repository

import (
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/gofiber/storage/memory/v2"
	"github.com/rs/zerolog"
	"github.com/valerius21/gitignore.lol/pkg/utils"
)

func init() {
	logger := utils.InitLogger()
	store := memory.New()

	DefaultRepository = Repository{
		logger: &logger,
		store:  store,
	}

	DefaultRepository.UpdateRepo()
}

type Repository struct {
	logger *zerolog.Logger
	store  *memory.Storage
}

var DefaultRepository Repository

// updateRepo clones the repo and updates the local copy
func (rr *Repository) UpdateRepo() {
	// check against cache
	cacheHit, err := rr.store.Get("repo")
	if err != nil {
		rr.logger.Fatal().Err(err).Msg("failed to check cache")
		return
	}

	if cacheHit != nil {
		rr.logger.Debug().Msg("cache hit")
		return
	}
	rr.logger.Debug().Msg("cache miss")
	info := utils.DefaultRepoInfo
	byteArr := []byte("fetched")
	rr.store.Set("repo", byteArr, 10*time.Minute)

	// if the repo already exists, pull the latest changes
	fileInfo, err := os.Stat(info.LocalPath)

	if err == nil && fileInfo.IsDir() {
		rr.logger.Debug().Msg("repo exists")
		r, err := git.PlainOpen(info.LocalPath)
		if err != nil {
			rr.logger.Fatal().Err(err).Msg("failed to open repo")
			return
		}

		w, err := r.Worktree()
		if err != nil {
			rr.logger.Fatal().Err(err).Msg("failed to get worktree")
			return
		}

		rr.logger.Debug().Msg("pulling repo")
		err = w.Pull(&git.PullOptions{RemoteName: "origin"})
		if err != nil {
			rr.logger.Fatal().Err(err).Msg("failed to pull repo")
			return
		}
		rr.logger.Debug().Msg("repo pulled")
	} else if os.IsNotExist(err) {
		// otherwise, clone the repo
		rr.logger.Debug().Msg("cloning repo")
		_, err = git.PlainClone(info.LocalPath, false, &git.CloneOptions{
			URL:   info.URL,
			Depth: 1,
		})

		if err != nil {
			rr.logger.Fatal().Err(err).Msg("failed to clone repo")
			return
		}
		rr.logger.Debug().Msg("repo cloned")
	} else {
		rr.logger.Fatal().Err(err).Msg("failed to check if repo exists")
		return
	}
}
