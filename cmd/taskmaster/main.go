package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"taskmaster/internal/config"
	"taskmaster/internal/daemon"
	"taskmaster/internal/logger"
	"taskmaster/internal/process"
	"taskmaster/internal/shell"
	"taskmaster/internal/web"
)

func main() {
	var configFile = flag.String("config", "configs/example.yml", "Path to configuration file")
	var webPort = flag.Int("web-port", 0, "Web server port (0 = disabled)")
	var daemonMode = flag.Bool("daemon", false, "Run as daemon (background)")
	var socketPath = flag.String("socket", "/tmp/taskmaster.sock", "Unix socket path for daemon")
	flag.Parse()

	// Initialize logger
	appLogger, err := logger.New("taskmaster.log")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Close()

	appLogger.Info("🚀 Starting Taskmaster...")

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		appLogger.Fatal("Failed to load config: %v", err)
	}

	appLogger.Info("✅ Configuration loaded from %s", *configFile)

	// Initialize process manager
	processManager := process.NewManager(cfg, appLogger)

	if *daemonMode {
		// MODO DAEMON
		appLogger.Info("🔧 Starting in daemon mode on socket %s", *socketPath)
		
		// Initialize daemon server
		daemonServer := daemon.NewServer(processManager, appLogger, *socketPath)
		
		// Initialize web server if requested
		if *webPort > 0 {
			webServer := web.NewServer(*webPort, processManager, appLogger)
			appLogger.SetBroadcaster(webServer.GetHub())
			processManager.SetStatusBroadcaster(webServer.GetHub())

			go func() {
				appLogger.Info("🌐 Starting web server on port %d", *webPort)
				if err := webServer.Start(); err != nil {
					appLogger.Error("Web server failed: %v", err)
				}
			}()
		}

		// Start processes marked as autostart
		if err := processManager.StartAutoStartProcesses(); err != nil {
			appLogger.Error("Failed to start some processes: %v", err)
		}

		// Start periodic status checking
		processManager.StartPeriodicStatusCheck()

		// Handle signals for daemon
		go handleDaemonSignals(processManager, appLogger, *configFile)

		// Start daemon server (blocks)
		if err := daemonServer.Start(); err != nil {
			appLogger.Fatal("Daemon server failed: %v", err)
		}

	} else {
		// MODO INTERACTIVO (original)
		appLogger.Info("🎮 Starting in interactive mode...")
		
		// Initialize web server only if port is specified
		if *webPort > 0 {
			webServer := web.NewServer(*webPort, processManager, appLogger)
			appLogger.SetBroadcaster(webServer.GetHub())
			processManager.SetStatusBroadcaster(webServer.GetHub())

			// Start web server in background
			go func() {
				appLogger.Info("🌐 Starting web server on port %d", *webPort)
				if err := webServer.Start(); err != nil {
					appLogger.Error("Web server failed: %v", err)
				}
			}()
		}

		// Start processes marked as autostart
		if err := processManager.StartAutoStartProcesses(); err != nil {
			appLogger.Error("Failed to start some processes: %v", err)
		}

		// Start periodic status checking
		processManager.StartPeriodicStatusCheck()

		// Handle SIGHUP for config reload
		go handleSignals(processManager, appLogger, *configFile)

		// Start interactive shell
		shellInstance := shell.New(processManager, appLogger)
		shellInstance.SetConfigFile(*configFile) // Pasar el archivo de configuración

		appLogger.Info("🎮 Starting interactive shell...")
		shellInstance.Run()
	}

	// Cleanup al salir
	appLogger.Info("🛑 Shutting down Taskmaster...")

	// Detener todos los procesos
	status := processManager.GetStatus()
	for programName := range status {
		if err := processManager.StopProgram(programName); err != nil {
			appLogger.Error("Error stopping program %s during shutdown: %v", programName, err)
		}
	}

	appLogger.Info("👋 Taskmaster shutdown complete")
}

func handleDaemonSignals(pm *process.Manager, logger *logger.Logger, configFile string) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	for sig := range sigChan {
		switch sig {
		case syscall.SIGHUP:
			logger.Info("📡 Received SIGHUP, reloading configuration...")
			if err := pm.ReloadConfig(configFile); err != nil {
				logger.Error("Failed to reload config: %v", err)
			} else {
				logger.Info("✅ Configuration reloaded via SIGHUP")
			}
		case syscall.SIGINT, syscall.SIGTERM:
			logger.Info("📡 Received shutdown signal, stopping daemon...")
			os.Exit(0)
		}
	}
}

func handleSignals(pm *process.Manager, logger *logger.Logger, configFile string) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	for sig := range sigChan {
		switch sig {
		case syscall.SIGHUP:
			logger.Info("📡 Received SIGHUP, reloading configuration...")
			if err := pm.ReloadConfig(configFile); err != nil {
				logger.Error("Failed to reload config: %v", err)
			} else {
				logger.Info("✅ Configuration reloaded via SIGHUP")
			}
		case syscall.SIGINT, syscall.SIGTERM:
			logger.Info("📡 Received shutdown signal, stopping all processes...")
			// El cleanup se hará en main() cuando termine el shell
			os.Exit(0)
		}
	}
}
