package process

import (
	"fmt"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// ProcessState representa el estado de un proceso
type ProcessState int

const (
	StateStopped ProcessState = iota
	StateStarting
	StateRunning
	StateFailed
)

// String convierte ProcessState a string
func (s ProcessState) String() string {
	switch s {
	case StateStopped:
		return "STOPPED"
	case StateStarting:
		return "STARTING"
	case StateRunning:
		return "RUNNING"
	case StateFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

// Process representa una instancia de proceso
type Process struct {
	Name      string
	Cmd       *exec.Cmd
	PID       int
	State     ProcessState
	StartTime time.Time
	Command   string
}

// Manager gestiona múltiples procesos
type Manager struct {
	processes map[string][]*Process
	mutex     sync.RWMutex
}

// NewManager crea un nuevo gestor de procesos
func NewManager() *Manager {
	return &Manager{
		processes: make(map[string][]*Process),
	}
}

// StartProgram inicia un programa (todas sus instancias)
func (m *Manager) StartProgram(name, command string, numProcs int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	fmt.Printf("🚀 Iniciando programa '%s' con %d procesos\n", name, numProcs)

	// Crear procesos
	processes := make([]*Process, numProcs)
	for i := 0; i < numProcs; i++ {
		processName := fmt.Sprintf("%s_%d", name, i)
		process := &Process{
			Name:    processName,
			State:   StateStarting,
			Command: command,
		}

		// Crear comando
		cmd := exec.Command("sh", "-c", command)
		process.Cmd = cmd

		// Iniciar proceso
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("error iniciando proceso %s: %w", processName, err)
		}

		process.PID = cmd.Process.Pid
		process.State = StateRunning
		process.StartTime = time.Now()

		processes[i] = process

		// Monitorear proceso en goroutine
		go m.monitorProcess(process)

		fmt.Printf("✅ Proceso %s iniciado (PID: %d)\n", processName, process.PID)
	}

	m.processes[name] = processes
	return nil
}

// StopProgram detiene un programa (todas sus instancias)
func (m *Manager) StopProgram(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	processes, exists := m.processes[name]
	if !exists {
		return fmt.Errorf("programa '%s' no encontrado", name)
	}

	fmt.Printf("🛑 Deteniendo programa '%s'\n", name)

	for _, process := range processes {
		if process.State == StateRunning {
			if err := process.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
				// Si SIGTERM falla, usar SIGKILL
				process.Cmd.Process.Kill()
			}
			process.State = StateStopped
			fmt.Printf("✅ Proceso %s detenido\n", process.Name)
		}
	}

	return nil
}

// RestartProgram reinicia un programa
func (m *Manager) RestartProgram(name string) error {
	fmt.Printf("🔄 Reiniciando programa '%s'\n", name)

	// Obtener configuración actual
	m.mutex.RLock()
	processes, exists := m.processes[name]
	if !exists {
		m.mutex.RUnlock()
		return fmt.Errorf("programa '%s' no encontrado", name)
	}
	
	command := processes[0].Command
	numProcs := len(processes)
	m.mutex.RUnlock()

	// Detener y luego iniciar
	if err := m.StopProgram(name); err != nil {
		return fmt.Errorf("error deteniendo programa: %w", err)
	}

	time.Sleep(2 * time.Second) // Esperar a que termine

	if err := m.StartProgram(name, command, numProcs); err != nil {
		return fmt.Errorf("error iniciando programa: %w", err)
	}

	return nil
}

// GetStatus obtiene el estado de todos los procesos
func (m *Manager) GetStatus() map[string][]*Process {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Crear copia para evitar race conditions
	status := make(map[string][]*Process)
	for name, processes := range m.processes {
		status[name] = make([]*Process, len(processes))
		copy(status[name], processes)
	}

	return status
}

// monitorProcess monitorea un proceso en background
func (m *Manager) monitorProcess(process *Process) {
	// Esperar a que termine el proceso
	err := process.Cmd.Wait()
	
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if err != nil {
		fmt.Printf("❌ Proceso %s terminó con error: %v\n", process.Name, err)
		process.State = StateFailed
	} else {
		fmt.Printf("✅ Proceso %s terminó normalmente\n", process.Name)
		process.State = StateStopped
	}
}

// IsProcessRunning verifica si un proceso sigue ejecutándose
func (m *Manager) IsProcessRunning(process *Process) bool {
	if process.Cmd == nil || process.Cmd.Process == nil {
		return false
	}

	// Enviar señal 0 para verificar si el proceso existe
	err := process.Cmd.Process.Signal(syscall.Signal(0))
	return err == nil
}