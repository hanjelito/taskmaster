package process

import (
	"fmt"
	"os/exec"
	"taskmaster/internal/config"
	"taskmaster/internal/logger"
	"taskmaster/pkg/signals"
	"time"
)

type Manager struct {
	processes map[string][]*ProcessInstance
	config    *config.Config
	logger    *logger.Logger
}

type ProcessInstance struct {
	Name      string
	Config    *ProcessConfig
	Cmd       *exec.Cmd
	PID       int
	State     ProcessState
	StartTime time.Time
	ExitCode  int
}

type ProcessState int

const (
	StateStopped ProcessState = iota
	StateStarting
	StateRunning
	StateFailed
)

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

type ProcessConfig struct {
	Cmd          string
	NumProcs     int
	AutoStart    bool
	AutoRestart  string
	ExitCodes    []int
	StartTime    int
	StartRetries int
	StopSignal   string
	StopTime     int
	Stdout       string
	Stderr       string
	Env          map[string]string
	WorkingDir   string
	Umask        string
}

func NewManager(cfg *config.Config, logger *logger.Logger) *Manager {
	return &Manager{
		processes: make(map[string][]*ProcessInstance),
		config:    cfg,
		logger:    logger,
	}
}

func (m *Manager) StartAutoStartProcesses() error {
	for name, program := range m.config.Programs {
		if program.AutoStart {
			if err := m.StartProgram(name); err != nil {
				m.logger.Error("Failed to start program %s: %v", name, err)
				continue
			}
		}
	}
	return nil
}

func (m *Manager) StartProgram(name string) error {
	program, exists := m.config.Programs[name]
	if !exists {
		return fmt.Errorf("program %s not found in configuration", name)
	}

	// Verificar si ya está ejecutándose
	if _, running := m.processes[name]; running {
		return fmt.Errorf("program %s is already running", name)
	}

	// Limpiar slice anterior
	m.processes[name] = nil

	// Crear e iniciar procesos reales
	for i := 0; i < program.NumProcs; i++ {
		processConfig := &ProcessConfig{
			Cmd:          program.Cmd,
			NumProcs:     program.NumProcs,
			AutoStart:    program.AutoStart,
			AutoRestart:  program.AutoRestart,
			ExitCodes:    program.ExitCodes,
			StartTime:    program.StartTime,
			StartRetries: program.StartRetries,
			StopSignal:   program.StopSignal,
			StopTime:     program.StopTime,
			Stdout:       program.Stdout,
			Stderr:       program.Stderr,
			Env:          program.Env,
			WorkingDir:   program.WorkingDir,
			Umask:        program.Umask,
		}

		instance := &ProcessInstance{
			Name:      fmt.Sprintf("%s_%d", name, i),
			Config:    processConfig,
			State:     StateStarting,
			StartTime: time.Now(),
		}

		// Crear comando real
		cmd := exec.Command("sh", "-c", program.Cmd)

		// Iniciar proceso
		if err := cmd.Start(); err != nil {
			m.logger.Error("Failed to start process %s: %v", instance.Name, err)
			instance.State = StateFailed
			continue
		}

		// Configurar instancia con proceso real
		instance.Cmd = cmd
		instance.PID = cmd.Process.Pid
		instance.State = StateRunning

		// Monitorear proceso en goroutine
		go m.monitorProcess(instance, name)

		m.processes[name] = append(m.processes[name], instance)
		m.logger.Info("Started process %s (PID: %d)", instance.Name, instance.PID)
	}

	return nil
}

func (m *Manager) monitorProcess(instance *ProcessInstance, programName string) {
	// Esperar a que el proceso termine
	err := instance.Cmd.Wait()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			instance.ExitCode = exitError.ExitCode()
		}
		m.logger.Error("Process %s exited with error: %v", instance.Name, err)
	} else {
		m.logger.Info("Process %s exited normally", instance.Name)
	}

	instance.State = StateStopped

	// Aquí podrías implementar lógica de restart automático
	// según la configuración autorestart
}

func (m *Manager) StopProgram(name string) error {
	instances, exists := m.processes[name]
	if !exists {
		return fmt.Errorf("program %s is not running", name)
	}

	for _, instance := range instances {
		if instance.Cmd != nil && instance.Cmd.Process != nil {
			// Usar señal configurable y timeout
			stopTimeout := time.Duration(instance.Config.StopTime) * time.Second

			m.logger.Info("Stopping process %s with signal %s (timeout: %ds)",
				instance.Name, instance.Config.StopSignal, instance.Config.StopTime)

			if err := signals.GracefulStop(instance.Cmd.Process, instance.Config.StopSignal, stopTimeout); err != nil {
				m.logger.Error("Failed to stop process %s gracefully: %v", instance.Name, err)
			} else {
				m.logger.Info("Process %s stopped gracefully", instance.Name)
			}
		}

		instance.State = StateStopped
		m.logger.Info("Stopped process %s (PID: %d)", instance.Name, instance.PID)
	}

	delete(m.processes, name)
	return nil
}

func (m *Manager) GetStatus() map[string][]*ProcessInstance {
	return m.processes
}

func (m *Manager) ReloadConfig(configFile string) error {
	newConfig, err := config.Load(configFile)
	if err != nil {
		return err
	}

	m.config = newConfig
	m.logger.Info("Configuration reloaded successfully")
	return nil
}
