package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func writeGitignoreFixture(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("failed to write fixture %s: %v", name, err)
	}

	return path
}

func TestGitRunner_ListFiles(t *testing.T) {
	tempDir := t.TempDir()
	goPath := writeGitignoreFixture(t, tempDir, "Go.gitignore", "# Go\n/bin/\n")
	nodePath := writeGitignoreFixture(t, tempDir, "Node.gitignore", "# Node\nnode_modules/\n")
	pythonPath := writeGitignoreFixture(t, tempDir, "Python.gitignore", "# Python\n__pycache__/\n")

	gr := NewGitRunner("", tempDir, 0)
	files, err := gr.ListFiles()
	if err != nil {
		t.Fatalf("ListFiles returned error: %v", err)
	}

	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d (%v)", len(files), files)
	}

	found := make(map[string]bool, len(files))
	for _, file := range files {
		found[file] = true
	}

	for _, expected := range []string{"go", "node", "python"} {
		if !found[expected] {
			t.Fatalf("expected %q in list, got %v", expected, files)
		}
	}

	if gr.langPaths["go"] != goPath {
		t.Fatalf("expected go path %q, got %q", goPath, gr.langPaths["go"])
	}
	if gr.langPaths["node"] != nodePath {
		t.Fatalf("expected node path %q, got %q", nodePath, gr.langPaths["node"])
	}
	if gr.langPaths["python"] != pythonPath {
		t.Fatalf("expected python path %q, got %q", pythonPath, gr.langPaths["python"])
	}
}

func TestGitRunner_GetFileContents(t *testing.T) {
	tempDir := t.TempDir()
	goContent := "# Go\n/bin/\n*.test\n"
	writeGitignoreFixture(t, tempDir, "Go.gitignore", goContent)

	gr := NewGitRunner("", tempDir, 0)
	_, err := gr.ListFiles()
	if err != nil {
		t.Fatalf("ListFiles returned error: %v", err)
	}

	content, err := gr.GetFileContents("go")
	if err != nil {
		t.Fatalf("GetFileContents(go) returned error: %v", err)
	}
	if content != goContent {
		t.Fatalf("expected content %q, got %q", goContent, content)
	}

	content, err = gr.GetFileContents("Go")
	if err != nil {
		t.Fatalf("GetFileContents(Go) returned error: %v", err)
	}
	if content != goContent {
		t.Fatalf("expected content %q for case-insensitive lookup, got %q", goContent, content)
	}
}

func TestGitRunner_GetFileContents_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	gr := NewGitRunner("", tempDir, 0)

	_, err := gr.GetFileContents("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestGitRunner_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	goContent := "# Go\n/bin/\n"
	writeGitignoreFixture(t, tempDir, "Go.gitignore", goContent)
	writeGitignoreFixture(t, tempDir, "Node.gitignore", "# Node\nnode_modules/\n")
	writeGitignoreFixture(t, tempDir, "Python.gitignore", "# Python\n__pycache__/\n")

	gr := NewGitRunner("", tempDir, 0)
	_, err := gr.ListFiles()
	if err != nil {
		t.Fatalf("initial ListFiles returned error: %v", err)
	}

	start := make(chan struct{})
	errCh := make(chan error, 20)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if _, err := gr.ListFiles(); err != nil {
				errCh <- err
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			content, err := gr.GetFileContents("go")
			if err != nil {
				errCh <- err
				return
			}
			if content != goContent {
				errCh <- fmt.Errorf("unexpected content: %q", content)
			}
		}()
	}

	close(start)
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent operation returned error: %v", err)
		}
	}
}

func TestGitRunner_ListFiles_Empty(t *testing.T) {
	tempDir := t.TempDir()
	gr := NewGitRunner("", tempDir, 0)

	files, err := gr.ListFiles()
	if err != nil {
		t.Fatalf("ListFiles returned error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected empty file list, got %v", files)
	}
	if len(gr.langPaths) != 0 {
		t.Fatalf("expected empty langPaths map, got %v", gr.langPaths)
	}
}
