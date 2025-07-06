package process

import (
	"os/exec"
	"sync"
	"taskmaster/internal/config"
	"taskmaster/internal/logger"
	"time"
)

// Manager gestiona múltiples procesos y sus instancias
type Manager struct {
	processes map[string][]*ProcessInstance
	config    *config.Config
	logger    *logger.Logger
	mutex     sync.RWMutex
}

// ProcessInstance representa una instancia específica de un proceso
type ProcessInstance struct {
	Name         string
	Config       *ProcessConfig
	Cmd          *exec.Cmd
	PID          int
	State        ProcessState
	StartTime    time.Time
	ExitCode     int
	RestartCount int
	StopChan     chan bool
	ManualStop   bool
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
