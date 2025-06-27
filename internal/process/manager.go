package process

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"taskmaster/internal/config"
	"taskmaster/internal/logger"
	"taskmaster/pkg/signals"
	"time"
)

type Manager struct {
	processes map[string][]*ProcessInstance
	config    *config.Config
	logger    *logger.Logger
	mutex     sync.RWMutex
}

type ProcessInstance struct {
	Name         string
	Config       *ProcessConfig
	Cmd          *exec.Cmd
	PID          int
	State        ProcessState
	StartTime    time.Time
	ExitCode     int
	RestartCount int
	StopChan     chan bool
	ManualStop   bool
}

type ProcessState int

const (
	StateStopped ProcessState = iota
	StateStarting
	StateRunning
	StateFailed
	StateRestarting
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
	case StateRestarting:
		return "RESTARTING"
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

// REEMPLAZAR el método StartProgram existente con esta versión:

func (m *Manager) StartProgram(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	program, exists := m.config.Programs[name]
	if !exists {
		return fmt.Errorf("program %s not found in configuration", name)
	}

	// Verificar si hay procesos activos usando el método auxiliar
	hasActive, activeCount := m.HasActiveProcesses(name)
	if hasActive {
		return fmt.Errorf("program %s has %d active processes running", name, activeCount)
	}

	// Limpiar automáticamente procesos muertos
	m.AutoCleanupProgram(name)

	// Crear configuración de proceso
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

	// Crear e iniciar procesos
	for i := 0; i < program.NumProcs; i++ {
		instance := &ProcessInstance{
			Name:      fmt.Sprintf("%s_%d", name, i),
			Config:    processConfig,
			State:     StateStarting,
			StartTime: time.Now(),
			StopChan:  make(chan bool, 1),
		}

		if err := m.startProcessInstance(instance, name); err != nil {
			m.logger.Error("Failed to start process %s: %v", instance.Name, err)
			instance.State = StateFailed
			continue
		}

		m.processes[name] = append(m.processes[name], instance)
		m.logger.Info("Started process %s (PID: %d)", instance.Name, instance.PID)
	}

	return nil
}

func (m *Manager) startProcessInstance(instance *ProcessInstance, programName string) error {
	// ✅ RESETEAR la bandera de parada manual al iniciar
	instance.ManualStop = false
	// Crear comando con umask si se especifica
	var cmd *exec.Cmd

	if instance.Config.Umask != "" {
		// Validar formato de umask
		if _, err := strconv.ParseUint(instance.Config.Umask, 8, 32); err != nil {
			m.logger.Error("Invalid umask format for %s: %s", instance.Name, instance.Config.Umask)
			// Crear comando sin umask
			cmd = exec.Command("sh", "-c", instance.Config.Cmd)
		} else {
			// Crear comando con umask usando shell wrapper
			wrappedCmd := fmt.Sprintf("umask %s; exec %s", instance.Config.Umask, instance.Config.Cmd)
			cmd = exec.Command("sh", "-c", wrappedCmd)
		}
	} else {
		// Crear comando normal
		cmd = exec.Command("sh", "-c", instance.Config.Cmd)
	}

	// Configurar entorno
	if len(instance.Config.Env) > 0 {
		env := os.Environ()
		for key, value := range instance.Config.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	// Configurar directorio de trabajo
	if instance.Config.WorkingDir != "" {
		cmd.Dir = instance.Config.WorkingDir
	}

	// Configurar redirección de stdout/stderr
	if instance.Config.Stdout != "" {
		if instance.Config.Stdout != "/dev/null" {
			if file, err := os.OpenFile(instance.Config.Stdout, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
				cmd.Stdout = file
			}
		}
	}

	if instance.Config.Stderr != "" {
		if instance.Config.Stderr != "/dev/null" {
			if file, err := os.OpenFile(instance.Config.Stderr, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
				cmd.Stderr = file
			}
		}
	}

	// Configurar atributos del proceso
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Crear nuevo grupo de procesos para mejor control
	}

	// Iniciar proceso
	if err := cmd.Start(); err != nil {
		return err
	}

	// Configurar instancia
	instance.Cmd = cmd
	instance.PID = cmd.Process.Pid
	instance.State = StateRunning

	// Monitorear proceso en goroutine
	go m.monitorProcess(instance, programName)

	return nil
}

func (m *Manager) monitorProcess(instance *ProcessInstance, programName string) {
	// Esperar tiempo de inicio para confirmar que el proceso se inició correctamente
	time.Sleep(time.Duration(instance.Config.StartTime) * time.Second)

	// Verificar si el proceso aún está en ejecución
	if instance.Cmd.ProcessState == nil {
		instance.State = StateRunning
		m.logger.Info("Process %s successfully started and running", instance.Name)
	}

	// Esperar a que el proceso termine
	err := instance.Cmd.Wait()
	exitCode := 0

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		instance.ExitCode = exitCode
		m.logger.Error("Process %s exited with code %d", instance.Name, exitCode)
	} else {
		m.logger.Info("Process %s exited normally", instance.Name)
	}

	// ✅ NUEVA LÓGICA: Si fue parada manual, no reiniciar
	if instance.ManualStop {
		m.logger.Info("Process %s was manually stopped, not restarting", instance.Name)
		instance.State = StateStopped
		return
	}

	// Determinar si se debe reiniciar (solo si NO fue parada manual)
	shouldRestart := m.shouldRestart(instance, exitCode)

	if shouldRestart && instance.RestartCount < instance.Config.StartRetries {
		m.logger.Info("Restarting process %s (attempt %d/%d)",
			instance.Name, instance.RestartCount+1, instance.Config.StartRetries)

		instance.State = StateRestarting
		instance.RestartCount++

		// Pequeña pausa antes de reiniciar
		time.Sleep(time.Second)

		if err := m.startProcessInstance(instance, programName); err != nil {
			m.logger.Error("Failed to restart process %s: %v", instance.Name, err)
			instance.State = StateFailed
		}
	} else {
		if instance.RestartCount >= instance.Config.StartRetries {
			m.logger.Error("Process %s failed too many times, giving up", instance.Name)
			instance.State = StateFailed
		} else {
			instance.State = StateStopped
		}
	}
}

func (m *Manager) shouldRestart(instance *ProcessInstance, exitCode int) bool {
	switch instance.Config.AutoRestart {
	case "always":
		return true
	case "never":
		return false
	case "unexpected":
		// Verificar si el código de salida es esperado
		for _, expected := range instance.Config.ExitCodes {
			if exitCode == expected {
				return false // Salida esperada, no reiniciar
			}
		}
		return true // Salida inesperada, reiniciar
	default:
		return false
	}
}

func (m *Manager) StopProgram(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	instances, exists := m.processes[name]
	if !exists {
		return fmt.Errorf("program %s is not running", name)
	}

	for _, instance := range instances {
		if instance.Cmd != nil && instance.Cmd.Process != nil {
			// ✅ MARCAR COMO PARADA MANUAL
			instance.ManualStop = true

			// Señalar al proceso que se detenga
			select {
			case instance.StopChan <- true:
			default:
			}

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
	}

	// ✅ NO eliminar del mapa, mantener como STOPPED
	// delete(m.processes, name) // REMOVER ESTA LÍNEA

	return nil
}

func (m *Manager) GetStatus() map[string][]*ProcessInstance {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Crear copia para evitar modificaciones concurrentes
	status := make(map[string][]*ProcessInstance)
	for name, instances := range m.processes {
		status[name] = make([]*ProcessInstance, len(instances))
		copy(status[name], instances)
	}
	return status
}

func (m *Manager) ReloadConfig(configFile string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Cargar nueva configuración
	newConfig, err := config.Load(configFile)
	if err != nil {
		return err
	}

	oldPrograms := make(map[string]config.Program)
	for name, program := range m.config.Programs {
		oldPrograms[name] = program
	}

	// Actualizar configuración
	m.config = newConfig

	// Procesar cambios
	for name, newProgram := range newConfig.Programs {
		oldProgram, existed := oldPrograms[name]

		if !existed {
			// Programa nuevo, iniciarlo si autostart está habilitado
			if newProgram.AutoStart {
				m.logger.Info("Starting new program %s", name)
				if err := m.startProgramUnsafe(name); err != nil {
					m.logger.Error("Failed to start new program %s: %v", name, err)
				}
			}
		} else {
			// Programa existente, verificar si cambió
			if !m.programsEqual(oldProgram, newProgram) {
				m.logger.Info("Program %s configuration changed, restarting", name)
				if err := m.stopProgramUnsafe(name); err != nil {
					m.logger.Error("Failed to stop program %s for restart: %v", name, err)
				}
				if newProgram.AutoStart {
					if err := m.startProgramUnsafe(name); err != nil {
						m.logger.Error("Failed to restart program %s: %v", name, err)
					}
				}
			}
		}
		delete(oldPrograms, name)
	}

	// Detener programas que ya no están en la configuración
	for name := range oldPrograms {
		m.logger.Info("Removing program %s (no longer in configuration)", name)
		if err := m.stopProgramUnsafe(name); err != nil {
			m.logger.Error("Failed to stop removed program %s: %v", name, err)
		}
	}

	m.logger.Info("Configuration reloaded successfully")
	return nil
}

func (m *Manager) startProgramUnsafe(name string) error {
	// Versión interna sin lock (ya tenemos el lock)
	program, exists := m.config.Programs[name]
	if !exists {
		return fmt.Errorf("program %s not found in configuration", name)
	}

	// Limpiar slice anterior
	m.processes[name] = nil

	// Crear configuración de proceso
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

	// Crear e iniciar procesos
	for i := 0; i < program.NumProcs; i++ {
		instance := &ProcessInstance{
			Name:      fmt.Sprintf("%s_%d", name, i),
			Config:    processConfig,
			State:     StateStarting,
			StartTime: time.Now(),
			StopChan:  make(chan bool, 1),
		}

		if err := m.startProcessInstance(instance, name); err != nil {
			m.logger.Error("Failed to start process %s: %v", instance.Name, err)
			instance.State = StateFailed
			continue
		}

		m.processes[name] = append(m.processes[name], instance)
		m.logger.Info("Started process %s (PID: %d)", instance.Name, instance.PID)
	}

	return nil
}

func (m *Manager) stopProgramUnsafe(name string) error {
	// Versión interna sin lock (ya tenemos el lock)
	instances, exists := m.processes[name]
	if !exists {
		return fmt.Errorf("program %s is not running", name)
	}

	for _, instance := range instances {
		if instance.Cmd != nil && instance.Cmd.Process != nil {
			select {
			case instance.StopChan <- true:
			default:
			}

			stopTimeout := time.Duration(instance.Config.StopTime) * time.Second

			if err := signals.GracefulStop(instance.Cmd.Process, instance.Config.StopSignal, stopTimeout); err != nil {
				m.logger.Error("Failed to stop process %s gracefully: %v", instance.Name, err)
			} else {
				m.logger.Info("Process %s stopped gracefully", instance.Name)
			}
		}

		instance.State = StateStopped
	}

	delete(m.processes, name)
	return nil
}

func (m *Manager) programsEqual(old, new config.Program) bool {
	// Comparar campos importantes que requieren reinicio
	return old.Cmd == new.Cmd &&
		old.NumProcs == new.NumProcs &&
		old.AutoRestart == new.AutoRestart &&
		old.StopSignal == new.StopSignal &&
		old.StopTime == new.StopTime &&
		old.StartTime == new.StartTime &&
		old.StartRetries == new.StartRetries &&
		old.Stdout == new.Stdout &&
		old.Stderr == new.Stderr &&
		old.WorkingDir == new.WorkingDir &&
		old.Umask == new.Umask &&
		m.slicesEqual(old.ExitCodes, new.ExitCodes) &&
		m.mapsEqual(old.Env, new.Env)
}

func (m *Manager) slicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func (m *Manager) mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func (m *Manager) CleanupDeadProcesses() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cleaned := 0
	for name, instances := range m.processes {
		activeInstances := []*ProcessInstance{}
		for _, instance := range instances {
			if instance.State == StateRunning || instance.State == StateStarting || instance.State == StateRestarting {
				activeInstances = append(activeInstances, instance)
			} else {
				cleaned++
			}
		}

		if len(activeInstances) == 0 {
			// Si no hay instancias activas, remover completamente del mapa
			delete(m.processes, name)
		} else {
			// Actualizar con solo las instancias activas
			m.processes[name] = activeInstances
		}
	}

	if cleaned > 0 {
		m.logger.Info("Cleaned up %d dead process instances", cleaned)
	}
}

func (m *Manager) CleanupProgram(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if instances, exists := m.processes[name]; exists {
		activeInstances := []*ProcessInstance{}
		cleaned := 0

		for _, instance := range instances {
			if instance.State == StateRunning || instance.State == StateStarting || instance.State == StateRestarting {
				activeInstances = append(activeInstances, instance)
			} else {
				cleaned++
			}
		}

		if len(activeInstances) == 0 {
			delete(m.processes, name)
		} else {
			m.processes[name] = activeInstances
		}

		if cleaned > 0 {
			m.logger.Info("Cleaned up %d dead instances for program %s", cleaned, name)
		}
	}
}

// HasActiveProcesses verifica si un programa tiene procesos activos
func (m *Manager) HasActiveProcesses(programName string) (bool, int) {
	instances, exists := m.processes[programName]
	if !exists {
		return false, 0
	}

	activeCount := 0
	for _, instance := range instances {
		if instance.State == StateRunning || instance.State == StateStarting || instance.State == StateRestarting {
			activeCount++
		}
	}

	return activeCount > 0, activeCount
}

// AutoCleanupProgram limpia automáticamente procesos muertos de un programa específico
func (m *Manager) AutoCleanupProgram(programName string) {
	instances, exists := m.processes[programName]
	if !exists {
		return
	}

	activeInstances := []*ProcessInstance{}
	cleanedCount := 0

	for _, instance := range instances {
		if instance.State == StateRunning || instance.State == StateStarting || instance.State == StateRestarting {
			activeInstances = append(activeInstances, instance)
		} else {
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		if len(activeInstances) == 0 {
			delete(m.processes, programName)
		} else {
			m.processes[programName] = activeInstances
		}
		m.logger.Info("Auto-cleaned %d dead instances for program %s", cleanedCount, programName)
	}
}
