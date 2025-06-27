package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"taskmaster/internal/config"
	"taskmaster/internal/logger"
	"taskmaster/internal/process"
	"taskmaster/internal/shell"
)

func main() {
	var configFile = flag.String("config", "configs/example.yml", "Path to configuration file")
	flag.Parse()

	// Initialize logger
	appLogger, err := logger.New("taskmaster.log")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Close()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		appLogger.Fatal("Failed to load config: %v", err)
	}

	// Initialize process manager
	processManager := process.NewManager(cfg, appLogger)

	// Start processes marked as autostart
	if err := processManager.StartAutoStartProcesses(); err != nil {
		appLogger.Error("Failed to start some processes: %v", err)
	}

	// Handle SIGHUP for config reload
	go handleSignals(processManager, appLogger, *configFile)

	// Start interactive shell
	shellInstance := shell.New(processManager, appLogger) // ← Cambio aquí
	shellInstance.Run()
}

func handleSignals(pm *process.Manager, logger *logger.Logger, configFile string) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)

	for sig := range sigChan {
		switch sig {
		case syscall.SIGHUP:
			logger.Info("Received SIGHUP, reloading configuration...")
			if err := pm.ReloadConfig(configFile); err != nil {
				logger.Error("Failed to reload config: %v", err)
			}
		}
	}
}
