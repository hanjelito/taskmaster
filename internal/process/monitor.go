package process

import (
	"os/exec"
	"syscall"
	"time"
)

// monitorProcess monitorea un proceso en ejecución
func (m *Manager) monitorProcess(instance *ProcessInstance, programName string) {
	time.Sleep(time.Duration(instance.Config.StartTime) * time.Second)

	if instance.Cmd.ProcessState == nil {
		instance.State = StateRunning
		m.logger.Info("Process %s successfully started and running", instance.Name)
		m.broadcastStatus()
	}

	err := instance.Cmd.Wait()
	exitCode := m.getExitCode(err)
	instance.ExitCode = exitCode

	if instance.ManualStop {
		m.logger.Info("Process %s stopped gracefully", instance.Name)
		instance.State = StateStopped
		m.broadcastStatus()
		return
	}

	m.handleProcessExit(instance, programName, exitCode, err)
}

// getExitCode extrae el código de salida de un error
func (m *Manager) getExitCode(err error) int {
	if err == nil {
		return 0
	}

	// Intentar obtener ExitError
	exitError, ok := err.(*exec.ExitError)
	if !ok {
		// Si no es ExitError, es un error desconocido
		return 1
	}

	// Obtener WaitStatus del sistema
	status, ok := exitError.Sys().(syscall.WaitStatus)
	if !ok {
		// Fallback: usar ExitCode() si no podemos obtener WaitStatus
		return exitError.ExitCode()
	}

	// Verificar si fue terminado por señal
	if status.Signaled() {
		// Proceso terminado por señal: 128 + número de señal
		sig := status.Signal()
		return 128 + int(sig)
	}

	// Proceso terminado normalmente
	return status.ExitStatus()
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
		m.finalizeProcess(instance, exitCode)
	}
}

// attemptRestart intenta reiniciar un proceso
func (m *Manager) attemptRestart(instance *ProcessInstance, programName string) {
	m.logger.Info("Restarting process %s (attempt %d/%d)",
		instance.Name, instance.RestartCount+1, instance.Config.StartRetries)

	instance.State = StateRestarting
	instance.RestartCount++
	m.broadcastStatus()

	time.Sleep(time.Second)

	if err := m.startProcessInstance(instance, programName); err != nil {
		m.logger.Error("Failed to restart process %s: %v", instance.Name, err)
		instance.State = StateFailed
		m.broadcastStatus()
	}
}

// finalizeProcess finaliza un proceso que no se puede reiniciar
func (m *Manager) finalizeProcess(instance *ProcessInstance, exitCode int) {
	if instance.RestartCount >= instance.Config.StartRetries {
		m.logger.Error("Process %s failed too many times, giving up", instance.Name)
		instance.State = StateFailed
		m.broadcastStatus()
		return
	}

	if instance.ManualStop {
		// Detenido intencionalmente por nuestro programa taskmaster (comando stop/restart)
		m.logger.Info("Process %s stopped by taskmaster", instance.Name)
		instance.State = StateStopped
	} else {
		// Terminación externa - distinguir entre natural y anómala
		if m.isExpectedExitCode(exitCode, instance.Config.ExitCodes) {
			// Exit code esperado = terminación natural/limpia
			m.logger.Info("Process %s terminated naturally with expected code %d", instance.Name, exitCode)
			instance.State = StateStopped // ← Cambio: STOPPED en lugar de FAILED
		} else {
			// Exit code inesperado = algo salió mal
			m.logger.Info("Process %s terminated with unexpected code %d", instance.Name, exitCode)
			instance.State = StateFailed
		}
	}
	m.broadcastStatus()
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
