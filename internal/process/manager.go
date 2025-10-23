package process

import (
	"fmt"
	"syscall"
	"taskmaster/internal/config"
	"taskmaster/internal/logger"
	"time"
)

// NewManager crea un nuevo gestor de procesos
func NewManager(cfg *config.Config, logger *logger.Logger) *Manager {
	return &Manager{
		processes: make(map[string][]*ProcessInstance),
		config:    cfg,
		logger:    logger,
	}
}

// SetStatusBroadcaster configura el broadcaster para enviar actualizaciones de estado
func (m *Manager) SetStatusBroadcaster(broadcaster StatusBroadcaster) {
	m.broadcaster = broadcaster
}

// broadcastStatus envía el estado actual de todos los procesos si hay un broadcaster configurado
func (m *Manager) broadcastStatus() {
	if m.broadcaster != nil {
		status := m.copyProcessMap()
		m.broadcaster.BroadcastStatus(status)
	}
}

// StartPeriodicStatusCheck inicia un monitoreo periódico del estado de los procesos
func (m *Manager) StartPeriodicStatusCheck() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			m.checkProcessStatus()
		}
	}()
}

// checkProcessStatus verifica el estado actual de todos los procesos
func (m *Manager) checkProcessStatus() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	statusChanged := false
	for programName, instances := range m.processes {
		for _, instance := range instances {
			if m.isProcessDead(instance) {
				statusChanged = true
				m.handleDeadProcess(instance, programName)
			}
		}
	}

	if statusChanged {
		m.broadcastStatus()
	}
}

// isProcessDead verifica si un proceso ha muerto
func (m *Manager) isProcessDead(instance *ProcessInstance) bool {
	return instance.State == StateRunning &&
		instance.Cmd != nil &&
		instance.Cmd.Process != nil &&
		instance.Cmd.Process.Signal(syscall.Signal(0)) != nil
}

// handleDeadProcess maneja un proceso que ha muerto externamente
func (m *Manager) handleDeadProcess(instance *ProcessInstance, programName string) {
	m.logger.Info("Process %s (PID %d) was killed externally", instance.Name, instance.PID)
	instance.State = StateStopped

	if m.shouldRestart(instance, 143) && instance.RestartCount < instance.Config.StartRetries {
		m.logger.Info("Attempting to restart externally killed process %s", instance.Name)
		go m.restartProcess(instance, programName)
	}
}

// restartProcess reinicia un proceso de forma asíncrona
func (m *Manager) restartProcess(instance *ProcessInstance, programName string) {
	time.Sleep(time.Second)
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if err := m.startProcessInstance(instance, programName); err != nil {
		m.logger.Error("Failed to restart process %s: %v", instance.Name, err)
		instance.State = StateFailed
		m.broadcastStatus()
	}
}

// StartAutoStartProcesses inicia todos los procesos marcados como autostart
func (m *Manager) StartAutoStartProcesses() error {
	var errors []string

	for name, program := range m.config.Programs {
		if program.AutoStart {
			if err := m.StartProgram(name); err != nil {
				m.logger.Error("Failed to start program %s: %v", name, err)
				errors = append(errors, fmt.Sprintf("%s: %v", name, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to start some programs: %v", errors)
	}
	return nil
}

// StartProgram inicia un programa específico
func (m *Manager) StartProgram(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.startProgramUnsafe(name)
}

// StopProgram detiene un programa específico
func (m *Manager) StopProgram(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.stopProgramUnsafe(name)
}

// GetStatus devuelve el estado actual de todos los procesos
func (m *Manager) GetStatus() map[string][]*ProcessInstance {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.copyProcessMap()
}

// ReloadConfig recarga la configuración y aplica los cambios
func (m *Manager) ReloadConfig(configFile string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	newConfig, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return m.applyConfigChanges(newConfig)
}

// CleanupDeadProcesses elimina todas las instancias de procesos muertos
func (m *Manager) CleanupDeadProcesses() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cleaned := 0
	for name := range m.processes {
		cleaned += m.cleanupProgramUnsafe(name)
	}

	if cleaned > 0 {
		m.logger.Info("Cleaned up %d dead process instances", cleaned)
	}
}

// CleanupProgram limpia las instancias muertas de un programa específico
func (m *Manager) CleanupProgram(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if cleaned := m.cleanupProgramUnsafe(name); cleaned > 0 {
		m.logger.Info("Cleaned up %d dead instances for program %s", cleaned, name)
	}
}

// HasActiveProcesses verifica si un programa tiene procesos activos
func (m *Manager) HasActiveProcesses(programName string) (bool, int) {
	instances, exists := m.processes[programName]
	if !exists {
		return false, 0
	}

	activeCount := m.countActiveInstances(instances)
	return activeCount > 0, activeCount
}

// AutoCleanupProgram limpia automáticamente procesos muertos de un programa específico
func (m *Manager) AutoCleanupProgram(programName string) {
	if cleaned := m.cleanupProgramUnsafe(programName); cleaned > 0 {
		m.logger.Info("Auto-cleaned %d dead instances for program %s", cleaned, programName)
	}
}
