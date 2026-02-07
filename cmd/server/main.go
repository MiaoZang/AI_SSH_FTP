package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"ssh-ftp-proxy/internal/config"
	"ssh-ftp-proxy/internal/logger"
	"ssh-ftp-proxy/internal/server"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "Path to config file")
	flag.Parse()

	// 1. Load Config
	absPath, _ := filepath.Abs(*configPath)
	if err := config.LoadConfig(absPath); err != nil {
		fmt.Printf("Failed to load config from %s: %v\n", absPath, err)
		os.Exit(1)
	}

	// 2. Init Logger
	if err := logger.InitLogger(config.GlobalConfig.Log.Level, config.GlobalConfig.Log.File); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Log.Info("Starting AI SSH/FTP Proxy Service")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 3. Start WS Server (Async)
	wsSrv := server.NewWSServer()
	go func() {
		if err := wsSrv.Run(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("WS Server failed", "error", err)
		}
	}()

	// 4. Start HTTP Server (Async)
	httpSrv := server.NewServer()
	go func() {
		if err := httpSrv.Run(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("HTTP Server failed", "error", err)
		}
	}()

	// 5. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Log.Info("Received signal, shutting down...", "signal", sig)
	case <-ctx.Done():
		logger.Log.Info("Context cancelled, shutting down...")
	}

	// Give 10 seconds for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown servers
	logger.Log.Info("Stopping servers...")

	// Note: gin.Engine doesn't expose the underlying http.Server by default
	// For full graceful shutdown, we would need to modify the server package
	// For now, we just log and exit gracefully
	_ = shutdownCtx

	logger.Log.Info("AI SSH/FTP Proxy Service stopped")
}
