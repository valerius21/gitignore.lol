package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
)

type GitRunner struct {
	origin        string
	LocalPath     string
	fetchInterval int
	langPaths     map[string]string
	mu            sync.RWMutex
}

func NewGitRunner(origin, path string, fetchInterval int) *GitRunner {
	return &GitRunner{
		origin:        origin,
		LocalPath:     path,
		fetchInterval: fetchInterval,
		langPaths:     make(map[string]string),
	}
}

// Init checks if the local path is a valid git repository.
// If not (missing, empty, or non-git directory), it clones fresh.
// If it is, it pulls the latest changes.
func (gr *GitRunner) Init() error {
	// Try to open as a valid git repository first
	if _, err := git.PlainOpen(gr.LocalPath); err == nil {
		// Repo exists and is valid — update it
		Logger.Info("Update", "origin", gr.origin)
		if err = gr.updateRepo(); err != nil {
			return err
		}
		_, err = gr.ListFiles()
		return err
	}

	// Not a valid git repo (directory missing, empty Docker volume, or corrupted) — clone fresh
	Logger.Info("Cloning", "origin", gr.origin, "path", gr.LocalPath)

	// Remove the path if it exists (handles empty Docker volume mount or partial state)
	if _, statErr := os.Stat(gr.LocalPath); statErr == nil {
		if err := os.RemoveAll(gr.LocalPath); err != nil {
			return fmt.Errorf("failed to remove non-repo directory: %w", err)
		}
	}

	if _, err := git.PlainClone(gr.LocalPath, false, &git.CloneOptions{
		URL: gr.origin,
	}); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	_, err := gr.ListFiles()
	return err
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
	newMap := make(map[string]string)
	for i, file := range files {
		fileNames[i] = strings.ToLower(filepath.Base(file))
		fileNames[i] = strings.ReplaceAll(fileNames[i], ".gitignore", "")
		newMap[fileNames[i]] = file
	}

	gr.mu.Lock()
	gr.langPaths = newMap
	gr.mu.Unlock()

	uniqueFiles := RemoveEmptyString(fileNames)
	uniqueFiles = RemoveDuplicates(uniqueFiles)
	return uniqueFiles, nil
}

func (gr *GitRunner) GetFileContents(name string) (string, error) {
	name = strings.ToLower(name)

	// Use the pre-populated langPaths map that was created in ListFiles()
	gr.mu.RLock()
	filePath, exists := gr.langPaths[name]
	gr.mu.RUnlock()
	if !exists {
		return "", fmt.Errorf("gitignore file for '%s' not found", name)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	return string(content), nil
}
