package process

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"taskmaster/pkg/signals"
	"time"
)

// startProcessInstance inicia una instancia específica de proceso
func (m *Manager) startProcessInstance(instance *ProcessInstance, programName string) error {
	instance.ManualStop = false

	cmd, err := m.createCommand(instance)
	if err != nil {
		return fmt.Errorf("failed to create command: %w", err)
	}

	// ahora puede fallar por privilege de-escalation
	if err := m.configureCommand(cmd, instance); err != nil {
		return fmt.Errorf("failed to configure command: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	instance.Cmd = cmd
	instance.PID = cmd.Process.Pid
	instance.State = StateRunning

	go m.monitorProcess(instance, programName)

	// Broadcast status update
	m.broadcastStatus()

	return nil
}

// createCommand crea el comando a ejecutar
func (m *Manager) createCommand(instance *ProcessInstance) (*exec.Cmd, error) {
	if instance.Config.Umask != "" {
		if err := m.validateUmask(instance.Config.Umask); err != nil {
			m.logger.Error("Invalid umask format for %s: %s", instance.Name, instance.Config.Umask)
			return exec.Command("sh", "-c", instance.Config.Cmd), nil
		}
		wrappedCmd := fmt.Sprintf("umask %s; exec %s", instance.Config.Umask, instance.Config.Cmd)
		return exec.Command("sh", "-c", wrappedCmd), nil
	}

	return exec.Command("sh", "-c", instance.Config.Cmd), nil
}

// validateUmask valida el formato del umask
func (m *Manager) validateUmask(umask string) error {
	_, err := strconv.ParseUint(umask, 8, 32)
	return err
}

// configureCommand configura el comando con ambiente, directorio y redirecciones
func (m *Manager) configureCommand(cmd *exec.Cmd, instance *ProcessInstance) error {
	m.configureEnvironment(cmd, instance.Config.Env)
	m.configureWorkingDir(cmd, instance.Config.WorkingDir)
	m.configureRedirections(cmd, instance)
	
	
	return m.configureProcessAttributes(cmd, instance.Config)
}

// configureEnvironment configura las variables de ambiente
func (m *Manager) configureEnvironment(cmd *exec.Cmd, env map[string]string) {
	if len(env) > 0 {
		cmd.Env = os.Environ()
		for key, value := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}
}

// configureWorkingDir configura el directorio de trabajo
func (m *Manager) configureWorkingDir(cmd *exec.Cmd, workingDir string) {
	if workingDir != "" {
		cmd.Dir = workingDir
	}
}

// configureRedirections configura las redirecciones de stdout y stderr
func (m *Manager) configureRedirections(cmd *exec.Cmd, instance *ProcessInstance) {
	stdout := instance.Config.Stdout
	stderr := instance.Config.Stderr

	if stdout != "" && stdout != "/dev/null" {
		if m.logBroadcaster != nil {
			// Usar TeeWriter para escribir a archivo Y WebSocket
			teeWriter, err := NewProcessTeeWriter(stdout, instance.Name, "stdout", m.logBroadcaster)
			if err != nil {
				m.logger.Error("Failed to create stdout TeeWriter for %s: %v", instance.Name, err)
				// Fallback: redirección normal a archivo
				if file, err := os.OpenFile(stdout, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
					cmd.Stdout = file
				}
			} else {
				cmd.Stdout = teeWriter
			}
		} else {
			// Sin broadcaster, redirección normal a archivo
			if file, err := os.OpenFile(stdout, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
				cmd.Stdout = file
			}
		}
	}

	// ✅ CONFIGURAR STDERR con TeeWriter  
	if stderr != "" && stderr != "/dev/null" {
		if m.logBroadcaster != nil {
			// Usar TeeWriter para escribir a archivo Y WebSocket
			teeWriter, err := NewProcessTeeWriter(stderr, instance.Name, "stderr", m.logBroadcaster)
			if err != nil {
				m.logger.Error("Failed to create stderr TeeWriter for %s: %v", instance.Name, err)
				// Fallback: redirección normal a archivo
				if file, err := os.OpenFile(stderr, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
					cmd.Stderr = file
				}
			} else {
				cmd.Stderr = teeWriter
			}
		} else {
			// Sin broadcaster, redirección normal a archivo
			if file, err := os.OpenFile(stderr, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
				cmd.Stderr = file
			}
		}
	}
}

// configureProcessAttributes configura los atributos del proceso
func (m *Manager) configureProcessAttributes(cmd *exec.Cmd, config *ProcessConfig) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// APLICAR PRIVILEGE DE-ESCALATION
	if config.User != "" {
		creds, err := LookupUser(config.User, config.Group)
		if err != nil {
			return fmt.Errorf("privilege de-escalation failed: %w", err)
		}

		if creds != nil {
			m.logger.Info("De-escalating privilege for process: user=%s(%d) group=%s(%d)", 
				config.User, creds.UID, getGroupName(config.Group, creds.GID), creds.GID)
			
			ApplyCredentials(cmd.SysProcAttr, creds)
		}
	}

	return nil
}

// getGroupName obtiene el nombre del grupo para logging
func getGroupName(groupName string, gid uint32) string {
	if groupName != "" {
		return groupName
	}
	return fmt.Sprintf("gid:%d", gid)
}

// stopProcessInstance detiene una instancia específica de proceso
func (m *Manager) stopProcessInstance(instance *ProcessInstance) bool {
	if instance.Cmd == nil || instance.Cmd.Process == nil {
		m.logger.Info("Process %s has no associated process to stop", instance.Name)
		return false
	}

	// Verificar si el proceso ya terminó
	if instance.Cmd.ProcessState != nil {
		m.logger.Info("Process %s already finished, no need to stop", instance.Name)
		return false
	}

	// Verificar si el proceso aún existe enviando señal 0
	if err := instance.Cmd.Process.Signal(syscall.Signal(0)); err != nil {
		m.logger.Info("Process %s no longer exists (PID %d)", instance.Name, instance.PID)
		return false
	}

	instance.ManualStop = true

	select {
	case instance.StopChan <- true:
	default:
	}

	stopTimeout := time.Duration(instance.Config.StopTime) * time.Second

	m.logger.Info("Stopping process %s with signal %s (timeout: %ds)",
		instance.Name, instance.Config.StopSignal, instance.Config.StopTime)

	if err := signals.GracefulStop(instance.Cmd.Process, instance.Config.StopSignal, stopTimeout); err != nil {
		m.logger.Error("Failed to stop process %s gracefully: %v", instance.Name, err)
		return false
	}

	instance.State = StateStopped
	
	// Broadcast status update
	m.broadcastStatus()
	
	return true
}
