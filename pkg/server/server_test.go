package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
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
