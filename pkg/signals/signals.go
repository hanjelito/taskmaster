package signals

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// SignalMap mapea nombres de señales a syscall.Signal
var SignalMap = map[string]os.Signal{
	"TERM": syscall.SIGTERM,
	"KILL": syscall.SIGKILL,
	"INT":  syscall.SIGINT,
	"HUP":  syscall.SIGHUP,
	"USR1": syscall.SIGUSR1,
	"USR2": syscall.SIGUSR2,
	"QUIT": syscall.SIGQUIT,
	"STOP": syscall.SIGSTOP,
	"CONT": syscall.SIGCONT,
}

// GetSignal convierte un string a os.Signal
func GetSignal(signalName string) (os.Signal, error) {
	if sig, exists := SignalMap[signalName]; exists {
		return sig, nil
	}
	return nil, fmt.Errorf("unknown signal: %s", signalName)
}

// SendSignal envía una señal a un proceso
func SendSignal(process *os.Process, signalName string) error {
	sig, err := GetSignal(signalName)
	if err != nil {
		return err
	}

	return process.Signal(sig)
}

// GracefulStop intenta parar un proceso gracefully y luego fuerza kill
func GracefulStop(process *os.Process, stopSignal string, timeout time.Duration) error {
	// Enviar señal de parada graceful
	if err := SendSignal(process, stopSignal); err != nil {
		return fmt.Errorf("failed to send %s: %v", stopSignal, err)
	}

	// Crear canal para monitorear si el proceso termina
	done := make(chan bool, 1)

	// Goroutine para verificar si el proceso aún existe
	go func() {
		for {
			// Enviar señal 0 para verificar si el proceso existe
			if err := process.Signal(syscall.Signal(0)); err != nil {
				// Proceso ya no existe
				done <- true
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Esperar el timeout o que el proceso termine
	select {
	case <-done:
		return nil // Proceso terminó gracefully
	case <-time.After(timeout):
		// Timeout alcanzado, forzar kill
		return process.Kill()
	}
}

// ValidSignals retorna una lista de señales válidas
func ValidSignals() []string {
	signals := make([]string, 0, len(SignalMap))
	for name := range SignalMap {
		signals = append(signals, name)
	}
	return signals
}

// IsValidSignal verifica si una señal es válida
func IsValidSignal(signalName string) bool {
	_, exists := SignalMap[signalName]
	return exists
}
