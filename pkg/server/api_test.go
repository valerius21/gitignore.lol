package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	lib "me.valerius/gitignore-lol/pkg/lib"
)

func setupTestApp(t *testing.T) (*fiber.App, map[string]string) {
	t.Helper()

	tempDir := t.TempDir()
	fixtures := map[string]string{
		"Go.gitignore":     "# Go\n*.exe\n*.test\n",
		"Node.gitignore":   "# Node\nnode_modules/\nnpm-debug.log*\n",
		"Python.gitignore": "# Python\n__pycache__/\n*.py[cod]\n",
	}

	for name, content := range fixtures {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write fixture %s: %v", name, err)
		}
	}

	gr := lib.NewGitRunner("", tempDir, 0)
	if _, err := gr.ListFiles(); err != nil {
		t.Fatalf("failed to list fixture templates: %v", err)
	}

	app := fiber.New()
	apiGroup := app.Group("/api")
	apiGroup.Get("/list", func(c fiber.Ctx) error {
		return listTemplates(c, gr)
	})
	apiGroup.Get("/*", func(c fiber.Ctx) error {
		return getTemplates(c, gr)
	})

	return app, map[string]string{
		"go":     fixtures["Go.gitignore"],
		"node":   fixtures["Node.gitignore"],
		"python": fixtures["Python.gitignore"],
	}
}

func performGetRequest(t *testing.T, app *fiber.App, path string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed for %s: %v", path, err)
	}

	return resp
}

func readResponseBody(t *testing.T, resp *http.Response) string {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	return string(body)
}

func TestListTemplates_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	resp := performGetRequest(t, app, "/api/list")
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	var parsed TemplateResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	expected := []string{"go", "node", "python"}
	for _, template := range expected {
		if !slices.Contains(parsed.Files, template) {
			t.Fatalf("expected template %q in list, got %v", template, parsed.Files)
		}
	}
}

func TestGetTemplates_SingleTemplate(t *testing.T) {
	app, fixtures := setupTestApp(t)

	resp := performGetRequest(t, app, "/api/go")
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	body := readResponseBody(t, resp)
	if !strings.Contains(body, fixtures["go"]) {
		t.Fatalf("expected Go template content in response, got %q", body)
	}
}

func TestGetTemplates_MultipleTemplates(t *testing.T) {
	app, fixtures := setupTestApp(t)

	resp := performGetRequest(t, app, "/api/go,node")
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	body := readResponseBody(t, resp)
	if !strings.Contains(body, fixtures["go"]) {
		t.Fatalf("expected Go template content in response, got %q", body)
	}
	if !strings.Contains(body, fixtures["node"]) {
		t.Fatalf("expected Node template content in response, got %q", body)
	}
}

func TestGetTemplates_NotFound(t *testing.T) {
	app, _ := setupTestApp(t)

	resp := performGetRequest(t, app, "/api/nonexistent")
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", fiber.StatusBadRequest, resp.StatusCode)
	}

	var parsed ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("failed to decode error response body: %v", err)
	}

	if parsed.Error != "Template 'nonexistent' not found" {
		t.Fatalf("unexpected error response: %q", parsed.Error)
	}
}

func TestGetTemplates_Deduplication(t *testing.T) {
	app, fixtures := setupTestApp(t)

	resp := performGetRequest(t, app, "/api/go,go")
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	body := readResponseBody(t, resp)
	if strings.Count(body, fixtures["go"]) != 1 {
		t.Fatalf("expected Go template content exactly once, got %q", body)
	}
}

func TestGetTemplates_OrderPreservation(t *testing.T) {
	app, fixtures := setupTestApp(t)

	resp := performGetRequest(t, app, "/api/node,go")
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	body := readResponseBody(t, resp)
	nodePos := strings.Index(body, fixtures["node"])
	goPos := strings.Index(body, fixtures["go"])

	if nodePos == -1 || goPos == -1 {
		t.Fatalf("expected both Node and Go template content in response, got %q", body)
	}

	if nodePos > goPos {
		t.Fatalf("expected Node content before Go content, got %q", body)
	}
}

func TestGetTemplates_URLEncodedList(t *testing.T) {
	app, fixtures := setupTestApp(t)

	resp := performGetRequest(t, app, "/api/go%2Cnode")
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	body := readResponseBody(t, resp)
	if !strings.Contains(body, fixtures["go"]) || !strings.Contains(body, fixtures["node"]) {
		t.Fatalf("expected both Go and Node template content in response, got %q", body)
	}
}
