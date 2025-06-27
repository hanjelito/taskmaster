package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Programs map[string]Program `yaml:"programs"`
}

type Program struct {
	Cmd          string            `yaml:"cmd"`
	NumProcs     int               `yaml:"numprocs"`
	AutoStart    bool              `yaml:"autostart"`
	AutoRestart  string            `yaml:"autorestart"` // always, never, unexpected
	ExitCodes    []int             `yaml:"exitcodes"`
	StartTime    int               `yaml:"starttime"`    // seconds to consider "successfully started"
	StartRetries int               `yaml:"startretries"` // max restart attempts
	StopSignal   string            `yaml:"stopsignal"`   // TERM, KILL, USR1, etc.
	StopTime     int               `yaml:"stoptime"`     // seconds to wait before KILL
	Stdout       string            `yaml:"stdout"`       // stdout redirection
	Stderr       string            `yaml:"stderr"`       // stderr redirection
	Env          map[string]string `yaml:"env"`          // environment variables
	WorkingDir   string            `yaml:"workingdir"`   // working directory
	Umask        string            `yaml:"umask"`        // umask for process
}

func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// Aplicar valores por defecto
	for name, program := range config.Programs {
		if program.NumProcs == 0 {
			program.NumProcs = 1
		}
		if program.StopSignal == "" {
			program.StopSignal = "TERM"
		}
		if program.StopTime == 0 {
			program.StopTime = 10
		}
		if program.StartTime == 0 {
			program.StartTime = 1
		}
		if program.StartRetries == 0 {
			program.StartRetries = 3
		}
		if program.AutoRestart == "" {
			program.AutoRestart = "unexpected"
		}
		if len(program.ExitCodes) == 0 {
			program.ExitCodes = []int{0}
		}
		if program.Umask == "" {
			program.Umask = "022"
		}

		// Actualizar el mapa con los valores por defecto
		config.Programs[name] = program
	}

	return &config, nil
}
