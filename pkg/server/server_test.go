package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v3"
	lib "me.valerius/gitignore-lol/pkg/lib"
)

func setupServerApp(corsOrigins string) *fiber.App {
	app := fiber.New()
	applyGlobalMiddleware(app, corsOrigins)
	registerHealthRoute(app)
	return app
}

func TestHealthEndpoint(t *testing.T) {
	app := setupServerApp("*")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	var parsed HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("failed to decode health response: %v", err)
	}

	if parsed.Status != "ok" {
		t.Fatalf("expected status %q, got %q", "ok", parsed.Status)
	}

	t.Logf("GET /health -> %d {\"status\":%q}", resp.StatusCode, parsed.Status)
}

func TestCORSMiddlewareAppliedGlobally(t *testing.T) {
	app := setupServerApp("*")
	app.Get("/probe", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set("Origin", "https://example.com")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("CORS request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected status %d, got %d", fiber.StatusNoContent, resp.StatusCode)
	}

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin %q, got %q", "*", got)
	}

	t.Logf("Access-Control-Allow-Origin: %s", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestRecoverMiddlewareCatchesPanics(t *testing.T) {
	app := setupServerApp("*")
	app.Get("/panic", func(c fiber.Ctx) error {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("panic request should be recovered, got error: %v", err)
	}

	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", fiber.StatusInternalServerError, resp.StatusCode)
	}
}

func TestRootTemplateRedirectEligible(t *testing.T) {
	tempDir := t.TempDir()
	fixtures := map[string]string{
		"Python.gitignore": "# Python\n__pycache__/\n",
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

	eligible := isRootTemplateRedirectEligible("Python", gr)
	if !eligible {
		t.Fatalf("expected 'Python' to be redirect-eligible")
	}
}

func TestUnknownRootPathNotRedirectEligible(t *testing.T) {
	tempDir := t.TempDir()
	fixtures := map[string]string{
		"Python.gitignore": "# Python\n__pycache__/\n",
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

	eligible := isRootTemplateRedirectEligible("UnknownTemplate", gr)
	if eligible {
		t.Fatalf("expected 'UnknownTemplate' to NOT be redirect-eligible")
	}
}

func TestRootTemplateRedirect(t *testing.T) {
	tempDir := t.TempDir()
	fixtures := map[string]string{
		"Python.gitignore": "# Python\n__pycache__/\n",
		"Go.gitignore":     "# Go\n*.exe\n",
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
	applyGlobalMiddleware(app, "*")
	registerHealthRoute(app)

	app.Get("/:template", func(c fiber.Ctx) error {
		template := c.Params("template")

		if template == "api" || template == "health" || template == "swagger" || template == "stats" {
			return c.Next()
		}

		eligible := isRootTemplateRedirectEligible(template, gr)
		if eligible {
			return c.Redirect().Status(fiber.StatusTemporaryRedirect).To("/api/" + template)
		}
		return c.Next()
	})

	tests := []struct {
		name         string
		path         string
		wantStatus   int
		wantLocation string
		description  string
	}{
		{
			name:         "Python template redirect",
			path:         "/Python",
			wantStatus:   fiber.StatusTemporaryRedirect,
			wantLocation: "/api/Python",
			description:  "GET /Python should return 307 redirect to /api/Python",
		},
		{
			name:         "Go template redirect",
			path:         "/Go",
			wantStatus:   fiber.StatusTemporaryRedirect,
			wantLocation: "/api/Go",
			description:  "GET /Go should return 307 redirect to /api/Go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}

			if tt.wantLocation != "" {
				location := resp.Header.Get("Location")
				if location != tt.wantLocation {
					t.Fatalf("expected Location header %q, got %q", tt.wantLocation, location)
				}
			}

			t.Logf("%s: status=%d, location=%s", tt.description, resp.StatusCode, resp.Header.Get("Location"))
		})
	}
}

func TestReservedPathsNoRedirect(t *testing.T) {
	tempDir := t.TempDir()
	fixtures := map[string]string{
		"Python.gitignore": "# Python\n__pycache__/\n",
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
	applyGlobalMiddleware(app, "*")
	registerHealthRoute(app)

	app.Get("/api/list", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"files": []string{"test"}})
	})

	app.Get("/:template", func(c fiber.Ctx) error {
		template := c.Params("template")

		if template == "api" || template == "health" || template == "swagger" || template == "stats" {
			return c.Next()
		}

		eligible := isRootTemplateRedirectEligible(template, gr)
		if eligible {
			return c.Redirect().Status(fiber.StatusTemporaryRedirect).To("/api/" + template)
		}
		return c.Next()
	})

	tests := []struct {
		name        string
		path        string
		wantStatus  int
		description string
	}{
		{
			name:        "health endpoint no redirect",
			path:        "/health",
			wantStatus:  fiber.StatusOK,
			description: "GET /health should return 200, not redirect",
		},
		{
			name:        "api list no redirect",
			path:        "/api/list",
			wantStatus:  fiber.StatusOK,
			description: "GET /api/list should return 200, not redirect",
		},
		{
			name:        "unknown path returns 404",
			path:        "/UnknownTemplate123",
			wantStatus:  fiber.StatusNotFound,
			description: "GET /UnknownTemplate123 should return 404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}

			t.Logf("%s: status=%d", tt.description, resp.StatusCode)
		})
	}
}

func TestRootTemplateRedirectPreservesQueryString(t *testing.T) {
	tempDir := t.TempDir()
	fixtures := map[string]string{
		"Python.gitignore": "# Python\n__pycache__/\n",
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

	app, err := newApp(gr, nil, nil, false, "*")
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/Python?foo=bar&baz=qux", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", fiber.StatusTemporaryRedirect, resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	wantLocation := "/api/Python?foo=bar&baz=qux"
	if location != wantLocation {
		t.Fatalf("expected Location header %q, got %q", wantLocation, location)
	}

	t.Logf("GET /Python?foo=bar&baz=qux -> %d Location: %s", resp.StatusCode, location)
}

func TestRootTemplateTrailingSlashNoRedirect(t *testing.T) {
	tempDir := t.TempDir()
	fixtures := map[string]string{
		"Python.gitignore": "# Python\n__pycache__/\n",
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

	app, err := newApp(gr, nil, nil, false, "*")
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/Python/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected status %d for trailing slash path, got %d", fiber.StatusNotFound, resp.StatusCode)
	}

	t.Logf("GET /Python/ -> %d (should be 404, not redirect)", resp.StatusCode)
}

func TestStatsEndpointNotRedirected(t *testing.T) {
	tempDir := t.TempDir()
	fixtures := map[string]string{
		"Python.gitignore": "# Python\n__pycache__/\n",
		"stats.gitignore":  "# stats\n*.log\n",
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

	app, err := newApp(gr, nil, nil, true, "*")
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status %d for /stats endpoint, got %d", fiber.StatusOK, resp.StatusCode)
	}

	body := make([]byte, resp.ContentLength)
	resp.Body.Read(body)
	if len(body) == 0 {
		t.Fatalf("expected non-empty response body from stats endpoint")
	}

	t.Logf("GET /stats -> %d (should be handled by stats endpoint, not redirected)", resp.StatusCode)
}
