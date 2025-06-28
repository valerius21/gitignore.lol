// Package main is the entry point for the gitignore.lol application
package main

import (
	"fmt"
	"time"

	"github.com/alecthomas/kong"

	"me.valerius/gitignore-lol/pkg/lib"
	"me.valerius/gitignore-lol/pkg/server"
)

// startUpdateTicker starts a goroutine that periodically updates the git repository.
func startUpdateTicker(gr *lib.GitRunner, intervalSeconds int) {
	interval := time.Duration(intervalSeconds) * time.Second
	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			lib.Logger.Info("Running scheduled git repository update...")
			err := gr.Init() // Init calls updateRepo if the repo exists
			if err != nil {
				lib.Logger.Error("Failed to update git repository in background", "error", err)
			} else {
				lib.Logger.Info("Git repository updated successfully in background.")
			}
		}
	}()
}

// @Summary      Main entry point
// @Description  Initializes the gitignore.lol application
func main() {
	// Parse command line flags
	_ = kong.Parse(&lib.CLI)

	gr := lib.NewGitRunner(lib.CLI.BaseRepository, lib.CLI.ClonePath, lib.CLI.UpdateInterval)
	lib.Logger.Info("CLI ARGS", "repository", lib.CLI.BaseRepository)

	err := gr.Init()
	if err != nil {
		lib.Logger.Error("Failed to initialize Git Repository", "error", err)
		panic(1)
	}

	lib.Logger.Info(fmt.Sprintf("Gitignores cloned to %s\n", gr.LocalPath))

	// Initialize rate limiter if enabled
	var rateLimiter *lib.MovingWindowLimiter
	var enhancedLimiter *lib.EnhancedRateLimiter

	if lib.CLI.EnableRateLimit {
		if lib.CLI.UseEnhancedLimiter {
			enhancedLimiter = lib.NewEnhancedRateLimiter(
				lib.CLI.RateLimit,
				lib.CLI.ErrorRateLimit,
				lib.CLI.RateWindow,
				lib.CLI.BlockMinutes,
				lib.CLI.MaxViolations,
				lib.CLI.RateCleanupMs,
			)
			lib.Logger.Info("Enhanced rate limiter initialized",
				"normal_limit", lib.CLI.RateLimit,
				"error_limit", lib.CLI.ErrorRateLimit,
				"window_seconds", lib.CLI.RateWindow,
				"block_minutes", lib.CLI.BlockMinutes,
				"max_violations", lib.CLI.MaxViolations,
				"cleanup_interval_ms", lib.CLI.RateCleanupMs)
		} else {
			rateLimiter = lib.NewMovingWindowLimiter(
				lib.CLI.RateLimit,
				lib.CLI.RateWindow,
				lib.CLI.RateCleanupMs,
			)
			lib.Logger.Info("Basic rate limiter initialized",
				"max_requests", lib.CLI.RateLimit,
				"window_seconds", lib.CLI.RateWindow,
				"cleanup_interval_ms", lib.CLI.RateCleanupMs)
		}
	}

	// Start the background update ticker
	startUpdateTicker(gr, lib.CLI.UpdateInterval)
	lib.Logger.Info("Started background repository update routine", "interval_seconds", lib.CLI.UpdateInterval)

	// Start the server with the configured port
	if err := server.Run(lib.CLI.Port, gr, rateLimiter, enhancedLimiter); err != nil {
		lib.Logger.Error("Server failed to start", "error", err)
	}

	// Gracefully stop rate limiters on shutdown
	if rateLimiter != nil {
		rateLimiter.Stop()
	}
	if enhancedLimiter != nil {
		enhancedLimiter.Stop()
	}
}
