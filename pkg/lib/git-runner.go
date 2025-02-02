package lib

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
)

type GitRunner struct {
	origin        string
	LocalPath     string
	fetchInterval int
}

func NewGitRunner(origin, path string, fetchInterval int) *GitRunner {
	return &GitRunner{
		origin:        origin,
		LocalPath:     path,
		fetchInterval: fetchInterval,
	}
}

// Init checks, if the local repository exists, if not clones it.
// If it does exist, it updates it.
func (gr *GitRunner) Init() error {
	// Check if directory exists, if not clone the repo
	if _, err := os.Stat(gr.LocalPath); os.IsNotExist(err) {
		Logger.Info("Cloning", "origin", gr.origin, "path", gr.LocalPath)
		_, err = git.PlainClone(gr.LocalPath, false, &git.CloneOptions{
			URL: gr.origin,
		})
		return err
	}

	Logger.Info("Update", "origin", gr.origin)
	return gr.updateRepo()
}

func (gr *GitRunner) updateRepo() error {
	// update repo

	repo, err := git.PlainOpen(gr.LocalPath)
	if err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = wt.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	// log latest commit hash that was pulled
	ref, err := repo.Head()
	if err != nil {
		return err
	}

	Logger.Info("Ref", "hash", ref.Hash().String())

	return nil
}

func (gr *GitRunner) ListFiles() ([]string, error) {
files, err := filepath.Glob(filepath.Join(gr.LocalPath, "*.gitignore"))
		if err != nil {
			return nil, err
		}

	fileNames := make([]string, len(files))
	for i, file := range files {
		fileNames[i] = strings.ToLower(filepath.Base(file))
		fileNames[i] = strings.ReplaceAll(fileNames[i], ".gitignore", "")
	}

	uniqueFiles := RemoveEmptyString(fileNames)
	uniqueFiles = RemoveDuplicates(uniqueFiles)
	return uniqueFiles, nil
}

