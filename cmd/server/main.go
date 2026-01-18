package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

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

	// 3. Start WS Server (Async)
	go func() {
		wsSrv := server.NewWSServer()
		if err := wsSrv.Run(); err != nil {
			logger.Log.Fatal("WS Server failed", "error", err)
		}
	}()

	// 4. Start HTTP Server
	srv := server.NewServer()
	if err := srv.Run(); err != nil {
		logger.Log.Fatal("HTTP Server failed", "error", err)
	}
}
