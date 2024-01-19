package repository

import (
	"os"
	"path/filepath"
	"strings"
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

const repoKey = "repo"

// updateRepo clones the repo and updates the local copy
func (rr *Repository) UpdateRepo() {
	// check against cache
	cacheHit, err := rr.store.Get(repoKey)
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
	rr.store.Reset()

	rr.store.Set(repoKey, []byte("fetched"), 10*time.Minute)

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

func (rr *Repository) GetMappedFileName(language string) (string, error) {
	lowerLanguage := strings.ToLower(language)

	// check against cache
	cacheHit, err := rr.store.Get(lowerLanguage)
	if err != nil {
		rr.logger.Fatal().Err(err).Msg("failed to check cache")
		return "", err
	}

	if cacheHit != nil {
		rr.logger.Debug().Msg("cache hit")
		return string(cacheHit), nil
	}

	rr.logger.Debug().Msg("cache miss")
	info := utils.DefaultRepoInfo
	// find all .gitignore files in the repo
	var gitignoreFiles []string // slice to hold matching filenames

	err = filepath.Walk(info.LocalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".gitignore") {
			gitignoreFiles = append(gitignoreFiles, path)
		}
		return nil
	})

	if err != nil {
		rr.logger.Err(err).Msgf("error walking the path %v\n", info.LocalPath)
	} else {
		rr.logger.Debug().Msg("Found .gitignore files:")
		for _, file := range gitignoreFiles {
			rr.logger.Debug().Msg(file)
			fileContents, err := os.ReadFile(file)
			if err != nil {
				rr.logger.Err(err).Msgf("failed to read file %s", file)
				return "", err
			}
			rr.store.Set(file, fileContents, 5*time.Minute)
			rr.logger.Debug().Msgf("Added file %s to cache: Size = %d", file, len(fileContents))
		}
	}

	// map the language to the file name
	// 1. the value is the file name
	// 2. the key is the language name, which consists of the file name without the .gitignore extension
	// 3. the key is lowercased
	mapping := make(map[string]string)
	for _, file := range gitignoreFiles {
		fileName := strings.TrimSuffix(filepath.Base(file), ".gitignore")
		lowerFileName := strings.ToLower(fileName)
		mapping[lowerFileName] = file
	}

	// add the mappings to the store
	for key, value := range mapping {
		rr.store.Set(key, []byte(value), 0)
	}

	// check if the language is in the store
	if value, err := rr.store.Get(lowerLanguage); err == nil {
		rr.logger.Debug().Msgf("Found file name mapping for %s: %s", lowerLanguage, value)
		return string(value), nil
	} else {
		rr.logger.Debug().Msgf("No file name mapping found for %s", lowerLanguage)
		return "", err
	}
}

// GetFileContent returns the content of the file with the given name
func (rr *Repository) GetFileContent(fileName, language string) ([]byte, error) {
	// check against cache
	cacheHit, err := rr.store.Get(fileName)
	if err != nil {
		rr.logger.Fatal().Err(err).Msg("failed to check cache")
		return nil, err
	}

	if cacheHit != nil {
		rr.logger.Debug().Msg("cache hit")
		return cacheHit, nil
	}

	rr.logger.Debug().Msg("cache miss")

	_, err = rr.GetMappedFileName(language)

	if err != nil {
		rr.logger.Err(err).Msg("failed to get mapped file name")
		return nil, err
	}

	// check if the file is in the store
	if value, err := rr.store.Get(fileName); err == nil {
		rr.logger.Debug().Msgf("Found file contents for %s", fileName)
		return value, nil
	} else {
		rr.logger.Debug().Msgf("No file contents found for %s", fileName)
		return nil, err
	}
}
