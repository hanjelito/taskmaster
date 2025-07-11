package shell

import (
	"bufio"
	"fmt"
	"mini-taskmaster/internal/config"
	"mini-taskmaster/internal/process"
	"os"
	"strings"
	"time"
)

// Shell representa la interfaz interactiva
type Shell struct {
	manager *process.Manager
	config  *config.Config
	scanner *bufio.Scanner
}

// New crea una nueva shell
func New(manager *process.Manager, config *config.Config) *Shell {
	return &Shell{
		manager: manager,
		config:  config,
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// Run ejecuta la shell interactiva
func (s *Shell) Run() {
	fmt.Println("🎮 Mini-Taskmaster Shell")
	fmt.Println("Escribe 'help' para ver comandos disponibles")
	
	// Iniciar procesos con autostart
	s.startAutoStartProcesses()

	for {
		fmt.Print("mini-taskmaster> ")
		
		if !s.scanner.Scan() {
			break
		}

		line := strings.TrimSpace(s.scanner.Text())
		if line == "" {
			continue
		}

		if s.executeCommand(line) {
			break // Comando exit
		}
	}

	fmt.Println("👋 Cerrando Mini-Taskmaster...")
}

// startAutoStartProcesses inicia procesos marcados como autostart
func (s *Shell) startAutoStartProcesses() {
	fmt.Println("🔄 Iniciando procesos con autostart...")
	
	for name, program := range s.config.Programs {
		if program.AutoStart {
			if err := s.manager.StartProgram(name, program.Cmd, program.NumProcs); err != nil {
				fmt.Printf("❌ Error iniciando %s: %v\n", name, err)
			}
		}
	}
}

// executeCommand ejecuta un comando de la shell
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
			fmt.Println("❌ Uso: start <nombre_programa>")
			return false
		}
		s.startProgram(args[0])
	case "stop":
		if len(args) == 0 {
			fmt.Println("❌ Uso: stop <nombre_programa>")
			return false
		}
		s.stopProgram(args[0])
	case "restart":
		if len(args) == 0 {
			fmt.Println("❌ Uso: restart <nombre_programa>")
			return false
		}
		s.restartProgram(args[0])
	case "exit", "quit":
		s.stopAllProcesses()
		return true
	default:
		fmt.Printf("❌ Comando desconocido: %s\n", command)
		fmt.Println("Escribe 'help' para ver comandos disponibles")
	}

	return false
}

// showHelp muestra la ayuda
func (s *Shell) showHelp() {
	fmt.Println("📚 Comandos disponibles:")
	fmt.Println("  help           - Muestra esta ayuda")
	fmt.Println("  status         - Muestra el estado de todos los procesos")
	fmt.Println("  start <nombre> - Inicia un programa")
	fmt.Println("  stop <nombre>  - Detiene un programa")
	fmt.Println("  restart <nombre> - Reinicia un programa")
	fmt.Println("  exit/quit      - Sale del programa")
}

// showStatus muestra el estado de todos los procesos
func (s *Shell) showStatus() {
	status := s.manager.GetStatus()
	
	if len(status) == 0 {
		fmt.Println("📋 No hay procesos ejecutándose")
		return
	}

	fmt.Printf("%-20s %-12s %-8s %-10s\n", "NOMBRE", "ESTADO", "PID", "TIEMPO")
	fmt.Println(strings.Repeat("-", 50))

	for _, processes := range status {
		for _, proc := range processes {
			uptime := "N/A"
			if proc.State == process.StateRunning {
				uptime = fmt.Sprintf("%.0fs", time.Since(proc.StartTime).Seconds())
			}

			pidStr := fmt.Sprintf("%d", proc.PID)
			if proc.State == process.StateStopped || proc.State == process.StateFailed {
				pidStr = "-"
			}

			fmt.Printf("%-20s %-12s %-8s %-10s\n",
				proc.Name,
				proc.State.String(),
				pidStr,
				uptime)
		}
	}
}

// startProgram inicia un programa
func (s *Shell) startProgram(name string) {
	program, exists := s.config.Programs[name]
	if !exists {
		fmt.Printf("❌ Programa '%s' no encontrado en la configuración\n", name)
		return
	}

	if err := s.manager.StartProgram(name, program.Cmd, program.NumProcs); err != nil {
		fmt.Printf("❌ Error iniciando %s: %v\n", name, err)
	}
}

// stopProgram detiene un programa
func (s *Shell) stopProgram(name string) {
	if err := s.manager.StopProgram(name); err != nil {
		fmt.Printf("❌ Error deteniendo %s: %v\n", name, err)
	}
}

// restartProgram reinicia un programa
func (s *Shell) restartProgram(name string) {
	if err := s.manager.RestartProgram(name); err != nil {
		fmt.Printf("❌ Error reiniciando %s: %v\n", name, err)
	}
}

// stopAllProcesses detiene todos los procesos al salir
func (s *Shell) stopAllProcesses() {
	fmt.Println("🛑 Deteniendo todos los procesos...")
	
	status := s.manager.GetStatus()
	for name := range status {
		if err := s.manager.StopProgram(name); err != nil {
			fmt.Printf("❌ Error deteniendo %s: %v\n", name, err)
		}
	}
	
	time.Sleep(2 * time.Second) // Esperar a que terminen
}