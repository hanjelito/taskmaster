package shell

import (
	"fmt"
	"strings"
	"taskmaster/internal/logger"
	"taskmaster/internal/process"
	"time"

	"github.com/chzyer/readline"
)

type Shell struct {
	manager *process.Manager
	logger  *logger.Logger
	rl      *readline.Instance
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

func (s *Shell) Run() {
	defer s.rl.Close()

	fmt.Println("Taskmaster shell started. Type 'help' for available commands.")

	for {
		line, err := s.rl.Readline()
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		s.executeCommand(line)
	}
}

func (s *Shell) executeCommand(line string) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return
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
			return
		}
		s.startProgram(args[0])
	case "stop":
		if len(args) == 0 {
			fmt.Println("Usage: stop <program_name>")
			return
		}
		s.stopProgram(args[0])
	case "restart":
		if len(args) == 0 {
			fmt.Println("Usage: restart <program_name>")
			return
		}
		s.restartProgram(args[0])
	case "reload":
		fmt.Println("Configuration reload not yet implemented")
	case "quit", "exit":
		fmt.Println("Goodbye!")
		return
	default:
		fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", command)
	}
}

func (s *Shell) showHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help     - Show this help message")
	fmt.Println("  status   - Show status of all programs")
	fmt.Println("  start    - Start a program")
	fmt.Println("  stop     - Stop a program")
	fmt.Println("  restart  - Restart a program")
	fmt.Println("  reload   - Reload configuration")
	fmt.Println("  quit/exit - Exit taskmaster")
}

func (s *Shell) showStatus() {
	status := s.manager.GetStatus()
	if len(status) == 0 {
		fmt.Println("No programs running")
		return
	}

	fmt.Printf("%-20s %-10s %-10s %-10s\n", "NAME", "STATE", "PID", "UPTIME")
	fmt.Println(strings.Repeat("-", 60))

	for _, instances := range status {
		for _, instance := range instances {
			uptime := "N/A"
			if !instance.StartTime.IsZero() {
				uptime = fmt.Sprintf("%.0fs", time.Since(instance.StartTime).Seconds())
			}
			fmt.Printf("%-20s %-10s %-10d %-10s\n",
				instance.Name,
				instance.State.String(),
				instance.PID,
				uptime)
		}
	}
}

func (s *Shell) startProgram(name string) {
	if err := s.manager.StartProgram(name); err != nil {
		fmt.Printf("Error starting program %s: %v\n", name, err)
	} else {
		fmt.Printf("Program %s started successfully\n", name)
	}
}

func (s *Shell) stopProgram(name string) {
	if err := s.manager.StopProgram(name); err != nil {
		fmt.Printf("Error stopping program %s: %v\n", name, err)
	} else {
		fmt.Printf("Program %s stopped successfully\n", name)
	}
}

func (s *Shell) restartProgram(name string) {
	fmt.Printf("Restarting program %s...\n", name)
	s.stopProgram(name)
	// Pequeña pausa para asegurar que el proceso se detenga
	time.Sleep(time.Second)
	s.startProgram(name)
}
