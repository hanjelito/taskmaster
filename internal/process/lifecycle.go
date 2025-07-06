package process

import (
	"fmt"
	"taskmaster/internal/config"
	"time"
)

// startProgramUnsafe inicia un programa sin bloquear (asume que ya se tiene el lock)
func (m *Manager) startProgramUnsafe(name string) error {
	program, exists := m.config.Programs[name]
	if !exists {
		return fmt.Errorf("program %s not found in configuration", name)
	}

	// Verificar procesos activos y limpiar si es necesario
	if hasActive, activeCount := m.HasActiveProcesses(name); hasActive {
		return fmt.Errorf("program %s has %d active processes running", name, activeCount)
	}

	m.AutoCleanupProgram(name)

	// Crear configuración de proceso
	processConfig := m.createProcessConfig(program)

	// Crear e iniciar procesos
	return m.createAndStartInstances(name, program.NumProcs, processConfig)
}

// stopProgramUnsafe detiene un programa sin bloquear (asume que ya se tiene el lock)
func (m *Manager) stopProgramUnsafe(name string) error {
	instances, exists := m.processes[name]
	if !exists {
		return fmt.Errorf("program %s is not running", name)
	}

	stoppedCount := 0
	for _, instance := range instances {
		if m.stopProcessInstance(instance) {
			stoppedCount++
		}
	}

	if stoppedCount > 0 {
		m.logger.Info("Successfully stopped %d process(es) for program %s", stoppedCount, name)
	}

	return nil
}

// createProcessConfig crea una configuración de proceso a partir de un programa
func (m *Manager) createProcessConfig(program config.Program) *ProcessConfig {
	return &ProcessConfig{
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
}

// createAndStartInstances crea e inicia múltiples instancias de un proceso
func (m *Manager) createAndStartInstances(name string, numProcs int, processConfig *ProcessConfig) error {
	var errors []string

	for i := 0; i < numProcs; i++ {
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
			errors = append(errors, fmt.Sprintf("%s: %v", instance.Name, err))
			continue
		}

		m.processes[name] = append(m.processes[name], instance)
		m.logger.Info("Started process %s (PID: %d)", instance.Name, instance.PID)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to start some instances: %v", errors)
	}
	return nil
}
