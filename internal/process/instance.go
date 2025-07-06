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

	m.configureCommand(cmd, instance)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	instance.Cmd = cmd
	instance.PID = cmd.Process.Pid
	instance.State = StateRunning

	go m.monitorProcess(instance, programName)

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
func (m *Manager) configureCommand(cmd *exec.Cmd, instance *ProcessInstance) {
	m.configureEnvironment(cmd, instance.Config.Env)
	m.configureWorkingDir(cmd, instance.Config.WorkingDir)
	m.configureRedirections(cmd, instance.Config.Stdout, instance.Config.Stderr)
	m.configureProcessAttributes(cmd)
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
func (m *Manager) configureRedirections(cmd *exec.Cmd, stdout, stderr string) {
	if stdout != "" && stdout != "/dev/null" {
		if file, err := os.OpenFile(stdout, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
			cmd.Stdout = file
		}
	}

	if stderr != "" && stderr != "/dev/null" {
		if file, err := os.OpenFile(stderr, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
			cmd.Stderr = file
		}
	}
}

// configureProcessAttributes configura los atributos del proceso
func (m *Manager) configureProcessAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// stopProcessInstance detiene una instancia específica de proceso
func (m *Manager) stopProcessInstance(instance *ProcessInstance) bool {
	if instance.Cmd == nil || instance.Cmd.Process == nil {
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
	return true
}
