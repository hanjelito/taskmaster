package shell

import (
	"fmt"
	"strings"
	"syscall"
	"taskmaster/internal/logger"
	"taskmaster/internal/process"
	"time"

	"github.com/chzyer/readline"
)

type Shell struct {
	manager    *process.Manager
	logger     *logger.Logger
	rl         *readline.Instance
	configFile string
}

func New(manager *process.Manager, logger *logger.Logger) *Shell {
	rl, err := readline.New("taskmaster> ")
	if err != nil {
		panic(err)
	}

	return &Shell{
		manager: manager,
		logger:  logger,
		rl:      rl,
	}
}

func (s *Shell) SetConfigFile(configFile string) {
	s.configFile = configFile
}

func (s *Shell) Run() {
	defer s.rl.Close()

	fmt.Println("🚀 Taskmaster shell started. Type 'help' for available commands.")

	for {
		line, err := s.rl.Readline()
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if s.executeCommand(line) {
			break // Comando quit/exit
		}
	}
}

func (s *Shell) executeCommand(line string) bool {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return false
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "help":
		s.showHelp()
	case "status":
		s.showStatus()
	case "start":
		if len(args) == 0 {
			fmt.Println("Usage: start <program_name>")
			return false
		}
		s.startProgram(args[0])
	case "stop":
		if len(args) == 0 {
			fmt.Println("Usage: stop <program_name>")
			return false
		}
		s.stopProgram(args[0])
	case "restart":
		if len(args) == 0 {
			fmt.Println("Usage: restart <program_name>")
			return false
		}
		s.restartProgram(args[0])
	case "reload":
		s.reloadConfig()
	case "clear":
		if len(args) == 0 {
			s.clearDeadProcesses()
		} else {
			s.clearSpecificProgram(args[0])
		}
	case "quit", "exit":
		fmt.Println("👋 Goodbye!")
		return true
	default:
		fmt.Printf("❌ Unknown command: %s. Type 'help' for available commands.\n", command)
	}
	return false
}

func (s *Shell) showHelp() {
	fmt.Println("📚 Available commands:")
	fmt.Println("  help     - Show this help message")
	fmt.Println("  status   - Show status of all programs")
	fmt.Println("  start    - Start a program")
	fmt.Println("  stop     - Stop a program")
	fmt.Println("  restart  - Restart a program")
	fmt.Println("  reload   - Reload configuration file")
	fmt.Println("  clear [program] - Clean process history (optional)")
	fmt.Println("  quit/exit - Exit taskmaster")
}

func (s *Shell) showStatus() {
	status := s.manager.GetStatus()
	if len(status) == 0 {
		fmt.Println("📋 No programs configured")
		return
	}

	fmt.Printf("%-20s %-12s %-8s %-10s %-8s\n", "NAME", "STATE", "PID", "UPTIME", "RESTARTS")
	fmt.Println(strings.Repeat("-", 70))

	for _, instances := range status {
		for _, instance := range instances {
			// Verificar si el proceso aún existe para estados RUNNING
			if instance.State == process.StateRunning && instance.Cmd != nil && instance.Cmd.Process != nil {
				if err := instance.Cmd.Process.Signal(syscall.Signal(0)); err != nil {
					// Proceso ya no existe, pero el estado no se actualizó
					fmt.Printf("%-20s %-12s %-8s %-10s %-8s (STALE)\n",
						instance.Name,
						"UNKNOWN",
						"-",
						"N/A",
						fmt.Sprintf("%d", instance.RestartCount))
					continue
				}
			}

			uptime := "N/A"
			pidStr := fmt.Sprintf("%d", instance.PID)

			// Solo mostrar uptime para procesos realmente corriendo
			if instance.State == process.StateRunning && !instance.StartTime.IsZero() {
				uptime = fmt.Sprintf("%.0fs", time.Since(instance.StartTime).Seconds())
			}

			// Para procesos terminados, no mostrar PID
			if instance.State == process.StateStopped || instance.State == process.StateFailed {
				pidStr = "-"
			}

			stateColor := s.getStateColor(instance.State)

			fmt.Printf("%-20s %s%-12s\033[0m %-8s %-10s %-8d\n",
				instance.Name,
				stateColor,
				instance.State.String(),
				pidStr,
				uptime,
				instance.RestartCount)
		}
	}
}

func (s *Shell) getStateColor(state process.ProcessState) string {
	stateStr := state.String()
	switch stateStr {
	case "RUNNING":
		return "\033[32m" // Verde
	case "FAILED":
		return "\033[31m" // Rojo
	case "STARTING", "RESTARTING":
		return "\033[33m" // Amarillo
	case "STOPPED":
		return "\033[90m" // Gris
	default:
		return ""
	}
}

func (s *Shell) startProgram(name string) {
	fmt.Printf("🚀 Starting program %s...\n", name)
	if err := s.manager.StartProgram(name); err != nil {
		fmt.Printf("❌ Error starting program %s: %v\n", name, err)
	} else {
		fmt.Printf("✅ Program %s started successfully\n", name)
	}
}

func (s *Shell) stopProgram(name string) {
	fmt.Printf("🛑 Stopping program %s...\n", name)
	if err := s.manager.StopProgram(name); err != nil {
		fmt.Printf("❌ Error stopping program %s: %v\n", name, err)
	} else {
		fmt.Printf("✅ Program %s stopped successfully\n", name)
	}
}

func (s *Shell) restartProgram(name string) {
	fmt.Printf("🔄 Restarting program %s...\n", name)

	// Parar el programa
	if err := s.manager.StopProgram(name); err != nil {
		fmt.Printf("❌ Error stopping program %s: %v\n", name, err)
		return
	}

	// Pequeña pausa para asegurar que el proceso se detenga
	time.Sleep(2 * time.Second)

	// Iniciar el programa
	if err := s.manager.StartProgram(name); err != nil {
		fmt.Printf("❌ Error starting program %s: %v\n", name, err)
	} else {
		fmt.Printf("✅ Program %s restarted successfully\n", name)
	}
}

func (s *Shell) clearDeadProcesses() {
	fmt.Println("🧹 Clearing dead processes from memory...")
	s.manager.CleanupDeadProcesses()
	fmt.Println("✅ Dead processes cleared")
	fmt.Println("ℹ️  Try starting your programs again")
}

func (s *Shell) clearSpecificProgram(name string) {
	fmt.Printf("🧹 Clearing dead processes for program %s...\n", name)
	s.manager.CleanupProgram(name)
	fmt.Printf("✅ Dead processes cleared for %s\n", name)
	fmt.Println("ℹ️  Try starting the program again")
}

func (s *Shell) reloadConfig() {
	fmt.Println("🔄 Reloading configuration...")

	if s.configFile == "" {
		fmt.Println("❌ No configuration file specified")
		return
	}

	if err := s.manager.ReloadConfig(s.configFile); err != nil {
		fmt.Printf("❌ Error reloading configuration: %v\n", err)
	} else {
		fmt.Println("✅ Configuration reloaded successfully")
		fmt.Println("ℹ️  Check status to see configuration changes")
	}
}
