package process

import (
	"fmt"
	"taskmaster/internal/config"
)

// applyConfigChanges aplica los cambios de configuración
func (m *Manager) applyConfigChanges(newConfig *config.Config) error {
	oldPrograms := make(map[string]config.Program)
	for name, program := range m.config.Programs {
		oldPrograms[name] = program
	}

	m.config = newConfig

	// Procesar programas nuevos y modificados
	for name, newProgram := range newConfig.Programs {
		if err := m.handleProgramChange(name, newProgram, oldPrograms); err != nil {
			m.logger.Error("Failed to handle program change %s: %v", name, err)
		}
		delete(oldPrograms, name)
	}

	// Detener programas eliminados
	for name := range oldPrograms {
		m.logger.Info("Removing program %s (no longer in configuration)", name)
		if err := m.stopProgramUnsafe(name); err != nil {
			m.logger.Error("Failed to stop removed program %s: %v", name, err)
		}
	}

	m.logger.Info("Configuration reloaded successfully")
	return nil
}

// handleProgramChange maneja los cambios en un programa específico
func (m *Manager) handleProgramChange(name string, newProgram config.Program, oldPrograms map[string]config.Program) error {
	oldProgram, existed := oldPrograms[name]

	if !existed {
		return m.handleNewProgram(name, newProgram)
	}

	return m.handleModifiedProgram(name, oldProgram, newProgram)
}

// handleNewProgram maneja un programa nuevo
func (m *Manager) handleNewProgram(name string, program config.Program) error {
	if program.AutoStart {
		m.logger.Info("Starting new program %s", name)
		return m.startProgramUnsafe(name)
	}
	return nil
}

// handleModifiedProgram maneja un programa modificado
func (m *Manager) handleModifiedProgram(name string, oldProgram, newProgram config.Program) error {
	if !m.programsEqual(oldProgram, newProgram) {
		m.logger.Info("Program %s configuration changed, restarting", name)

		if err := m.stopProgramUnsafe(name); err != nil {
			return fmt.Errorf("failed to stop program for restart: %w", err)
		}

		if newProgram.AutoStart {
			return m.startProgramUnsafe(name)
		}
	}
	return nil
}

// programsEqual compara dos configuraciones de programa
func (m *Manager) programsEqual(old, new config.Program) bool {
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

// slicesEqual compara dos slices de enteros
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

// mapsEqual compara dos mapas string-string
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
