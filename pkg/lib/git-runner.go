package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
)

type GitRunner struct {
	origin        string
	LocalPath     string
	fetchInterval int
	langPaths     map[string]string
}

func NewGitRunner(origin, path string, fetchInterval int) *GitRunner {
	return &GitRunner{
		origin:        origin,
		LocalPath:     path,
		fetchInterval: fetchInterval,
		langPaths:     make(map[string]string),
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
		gr.langPaths[fileNames[i]] = file
	}

	uniqueFiles := RemoveEmptyString(fileNames)
	uniqueFiles = RemoveDuplicates(uniqueFiles)
	return uniqueFiles, nil
}

func (gr *GitRunner) indexFiles() error {
	return nil
}

func (gr *GitRunner) GetFileContents(name string) (string, error) {
	pattern := filepath.Join(gr.LocalPath, "**/*.gitignore")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no matches found for pattern: %s", pattern)
	}

	name = strings.ToLower(name)
	for _, file := range files {
		baseName := strings.ToLower(filepath.Base(file))
		baseName = strings.ReplaceAll(baseName, ".gitignore", "")
		if baseName == name {
			content, err := os.ReadFile(file)
			if err != nil {
				return "", err
			}
			return string(content), nil
		}
	}

	return "", fmt.Errorf("file not found: %s", name)
}
