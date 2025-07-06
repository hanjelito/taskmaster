package process

import (
	"os/exec"
	"time"
)

// monitorProcess monitorea un proceso en ejecución
func (m *Manager) monitorProcess(instance *ProcessInstance, programName string) {
	time.Sleep(time.Duration(instance.Config.StartTime) * time.Second)

	if instance.Cmd.ProcessState == nil {
		instance.State = StateRunning
		m.logger.Info("Process %s successfully started and running", instance.Name)
	}

	err := instance.Cmd.Wait()
	exitCode := m.getExitCode(err)
	instance.ExitCode = exitCode

	if instance.ManualStop {
		m.logger.Info("Process %s stopped gracefully", instance.Name)
		instance.State = StateStopped
		return
	}

	m.handleProcessExit(instance, programName, exitCode, err)
}

// getExitCode extrae el código de salida de un error
func (m *Manager) getExitCode(err error) int {
	if err == nil {
		return 0
	}

	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode()
	}

	return 1
}

// handleProcessExit maneja la salida de un proceso
func (m *Manager) handleProcessExit(instance *ProcessInstance, programName string, exitCode int, err error) {
	if err != nil {
		m.logger.Error("Process %s exited with code %d", instance.Name, exitCode)
	} else {
		m.logger.Info("Process %s exited normally", instance.Name)
	}

	if m.shouldRestart(instance, exitCode) && instance.RestartCount < instance.Config.StartRetries {
		m.attemptRestart(instance, programName)
	} else {
		m.finalizeProcess(instance)
	}
}

// attemptRestart intenta reiniciar un proceso
func (m *Manager) attemptRestart(instance *ProcessInstance, programName string) {
	m.logger.Info("Restarting process %s (attempt %d/%d)",
		instance.Name, instance.RestartCount+1, instance.Config.StartRetries)

	instance.State = StateRestarting
	instance.RestartCount++

	time.Sleep(time.Second)

	if err := m.startProcessInstance(instance, programName); err != nil {
		m.logger.Error("Failed to restart process %s: %v", instance.Name, err)
		instance.State = StateFailed
	}
}

// finalizeProcess finaliza un proceso que no se puede reiniciar
func (m *Manager) finalizeProcess(instance *ProcessInstance) {
	if instance.RestartCount >= instance.Config.StartRetries {
		m.logger.Error("Process %s failed too many times, giving up", instance.Name)
		instance.State = StateFailed
	} else {
		instance.State = StateStopped
	}
}

// shouldRestart determina si un proceso debe reiniciarse
func (m *Manager) shouldRestart(instance *ProcessInstance, exitCode int) bool {
	switch instance.Config.AutoRestart {
	case "always":
		return true
	case "never":
		return false
	case "unexpected":
		return !m.isExpectedExitCode(exitCode, instance.Config.ExitCodes)
	default:
		return false
	}
}

// isExpectedExitCode verifica si un código de salida es esperado
func (m *Manager) isExpectedExitCode(exitCode int, expectedCodes []int) bool {
	for _, expected := range expectedCodes {
		if exitCode == expected {
			return true
		}
	}
	return false
}
