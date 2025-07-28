package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"taskmaster/internal/daemon"
	"taskmaster/internal/process"
	"time"

	"github.com/chzyer/readline"
)

func main() {
	var socketPath = flag.String("socket", "/tmp/taskmaster.sock", "Unix socket path to daemon")
	var command = flag.String("cmd", "", "Command to execute (non-interactive)")
	var program = flag.String("program", "", "Program name for command")
	flag.Parse()

	client := daemon.NewClient(*socketPath)

	if *command != "" {
		// Modo no interactivo
		if err := executeCommand(client, *command, *program); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Modo interactivo
	runInteractiveShell(client)
}

func executeCommand(client *daemon.Client, cmd, program string) error {
	switch cmd {
	case "status":
		return showStatus(client)
	case "start":
		if program == "" {
			return fmt.Errorf("program name required for start command")
		}
		return client.StartProgram(program)
	case "stop":
		if program == "" {
			return fmt.Errorf("program name required for stop command")
		}
		return client.StopProgram(program)
	case "restart":
		if program == "" {
			return fmt.Errorf("program name required for restart command")
		}
		return client.RestartProgram(program)
	case "reload":
		return client.ReloadConfig()
	case "shutdown":
		return client.Shutdown()
	case "attach":
		if program == "" {
			return fmt.Errorf("instance name required for attach command")
		}
		return client.AttachToProcess(program)
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func runInteractiveShell(client *daemon.Client) {
	rl, err := readline.New("taskmasterctl> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	fmt.Println("🚀 Taskmaster Control Shell. Type 'help' for commands.")

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if executeShellCommand(client, line) {
			break // quit/exit
		}
	}
}

func executeShellCommand(client *daemon.Client, line string) bool {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return false
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "help":
		showHelp()
	case "status":
		if err := showStatus(client); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		}
	case "start":
		if len(args) == 0 {
			fmt.Println("Usage: start <program_name>")
			return false
		}
		if err := client.StartProgram(args[0]); err != nil {
			fmt.Printf("❌ Error starting %s: %v\n", args[0], err)
		} else {
			fmt.Printf("✅ Program %s started successfully\n", args[0])
		}
	case "stop":
		if len(args) == 0 {
			fmt.Println("Usage: stop <program_name>")
			return false
		}
		if err := client.StopProgram(args[0]); err != nil {
			fmt.Printf("❌ Error stopping %s: %v\n", args[0], err)
		} else {
			fmt.Printf("✅ Program %s stopped successfully\n", args[0])
		}
	case "restart":
		if len(args) == 0 {
			fmt.Println("Usage: restart <program_name>")
			return false
		}
		if err := client.RestartProgram(args[0]); err != nil {
			fmt.Printf("❌ Error restarting %s: %v\n", args[0], err)
		} else {
			fmt.Printf("✅ Program %s restarted successfully\n", args[0])
		}
	case "attach":
		if len(args) == 0 {
			fmt.Println("Usage: attach <instance_name>")
			fmt.Println("Example: attach logger_program_0")
			return false
		}
		fmt.Printf("🔗 Attaching to %s...\n", args[0])
		if err := client.AttachToProcess(args[0]); err != nil {
			fmt.Printf("❌ Error attaching to %s: %v\n", args[0], err)
		}
		// Después del attach, volver al prompt normal
		fmt.Printf("\n🎮 Back to taskmasterctl shell\n")
	case "reload":
		if err := client.ReloadConfig(); err != nil {
			fmt.Printf("❌ Error reloading config: %v\n", err)
		} else {
			fmt.Println("✅ Configuration reloaded successfully")
		}
	case "shutdown":
		if err := client.Shutdown(); err != nil {
			fmt.Printf("❌ Error shutting down daemon: %v\n", err)
		} else {
			fmt.Println("✅ Daemon shutdown requested")
		}
		return true
	case "quit", "exit":
		fmt.Println("👋 Goodbye!")
		return true
	default:
		fmt.Printf("❌ Unknown command: %s. Type 'help' for available commands.\n", command)
	}
	return false
}

func showHelp() {
	fmt.Println("📚 Available commands:")
	fmt.Println("  help        - Show this help message")
	fmt.Println("  status      - Show status of all programs")
	fmt.Println("  start <prog> - Start a program")
	fmt.Println("  stop <prog>  - Stop a program")
	fmt.Println("  restart <prog> - Restart a program")
	fmt.Println("  attach <instance> - Attach to process output (Ctrl+C to detach)")
	fmt.Println("  reload      - Reload daemon configuration")
	fmt.Println("  shutdown    - Shutdown daemon")
	fmt.Println("  quit/exit   - Exit control shell")
	fmt.Println("")
	fmt.Println("💡 Examples:")
	fmt.Println("  attach logger_program_0   # Attach to specific instance")
	fmt.Println("  attach web_server_0       # See web server logs in real-time")
}

func showStatus(client *daemon.Client) error {
	status, err := client.GetStatus()
	if err != nil {
		return err
	}

	if len(status) == 0 {
		fmt.Println("📋 No programs configured")
		return nil
	}

	fmt.Printf("%-20s %-12s %-8s %-10s %-8s\n", "NAME", "STATE", "PID", "UPTIME", "RESTARTS")
	fmt.Println(strings.Repeat("-", 70))

	for _, instances := range status {
		for _, instance := range instances {
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

			stateColor := getStateColor(instance.State)

			fmt.Printf("%-20s %s%-12s\033[0m %-8s %-10s %-8d\n",
				instance.Name,
				stateColor,
				instance.State.String(),
				pidStr,
				uptime,
				instance.RestartCount)
		}
	}

	return nil
}

func getStateColor(state process.ProcessState) string {
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