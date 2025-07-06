package process

import (
	"fmt"
	"taskmaster/internal/config"
	"taskmaster/internal/logger"
)

// NewManager crea un nuevo gestor de procesos
func NewManager(cfg *config.Config, logger *logger.Logger) *Manager {
	return &Manager{
		processes: make(map[string][]*ProcessInstance),
		config:    cfg,
		logger:    logger,
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
