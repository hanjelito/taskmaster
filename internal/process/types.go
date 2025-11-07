package process

import (
	"os/exec"
	"sync"
	"taskmaster/internal/config"
	"taskmaster/internal/logger"
	"time"
)

// StatusBroadcaster es la interfaz para enviar actualizaciones de estado
type StatusBroadcaster interface {
	BroadcastStatus(status interface{})
}

// Manager gestiona múltiples procesos y sus instancias
type Manager struct {
	processes   map[string][]*ProcessInstance
	config      *config.Config
	logger      *logger.Logger
	mutex       sync.RWMutex
	broadcaster StatusBroadcaster
}

// ProcessInstance representa una instancia específica de un proceso
type ProcessInstance struct {
	Name         string         `json:"name"`          // Nombre del programa
	Config       *ProcessConfig `json:"-"`             // Configuración del proceso
	Cmd          *exec.Cmd      `json:"-"`             // Comando en ejecución
	PID          int            `json:"pid"`           // ID del proceso
	State        ProcessState   `json:"state"`         // Estado del proceso
	StartTime    time.Time      `json:"start_time"`    // Hora de inicio
	ExitCode     int            `json:"exit_code"`     // Código de salida
	RestartCount int            `json:"restart_count"` // Contador de reinicios
	StopChan     chan bool      `json:"-"`             // Canal para detener el proceso
	ManualStop   bool           `json:"manual_stop"`   // Indica si el proceso fue detenido manualmente
}

// ProcessState representa el estado actual de un proceso
type ProcessState int

const (
	StateStopped ProcessState = iota
	StateStarting
	StateRunning
	StateFailed
	StateRestarting
)

// String convierte ProcessState a string legible
func (s ProcessState) String() string {
	states := map[ProcessState]string{
		StateStopped:    "STOPPED",
		StateStarting:   "STARTING",
		StateRunning:    "RUNNING",
		StateFailed:     "FAILED",
		StateRestarting: "RESTARTING",
	}
	if state, exists := states[s]; exists {
		return state
	}
	return "UNKNOWN"
}

// MarshalJSON implements json.Marshaler interface to serialize ProcessState as string
func (s ProcessState) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}

// ProcessConfig contiene la configuración de un proceso
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
