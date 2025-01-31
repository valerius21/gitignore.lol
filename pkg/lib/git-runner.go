package lib

import (
	"os"

	"github.com/go-git/go-git/v5"
)

type GitRunner struct {
	origin        string
	localPath     string
	fetchInterval int
}

func New(origin, path string, fetchInterval int) *GitRunner {
	return &GitRunner{
		origin:        origin,
		localPath:     path,
		fetchInterval: fetchInterval,
	}
}

// Init checks, if the local repository exists, if not clones it.
// If it does exist, it updates it.
func (gr *GitRunner) Init() error {
	// Check if directory exists, if not clone the repo
	if _, err := os.Stat(gr.localPath); os.IsNotExist(err) {
		// TODO: log
		_, err = git.PlainClone(gr.localPath, false, &git.CloneOptions{
			URL: gr.origin,
		})

		if err != nil {
			return err
		}
		return nil
	}

	// update repo

	repo, err := git.PlainOpen(gr.localPath)
	if err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = wt.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil {
		return err
	}

	// log latest commit hash that was pulled
	_, err = repo.Head()
	if err != nil {
		return err
	}

	// TODO: log ref

	return nil
}
