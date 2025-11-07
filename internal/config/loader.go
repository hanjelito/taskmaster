package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

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
		if program.AutoRestart == "" {
			program.AutoRestart = "unexpected"
		}
		if len(program.ExitCodes) == 0 {
			program.ExitCodes = []int{0}
		}
		if program.StartTime == 0 {
			program.StartTime = 1
		}
		if program.StartRetries == 0 {
			program.StartRetries = 3
		}
		if program.StopSignal == "" {
			program.StopSignal = "TERM"
		}
		if program.StopTime == 0 {
			program.StopTime = 10
		}
		if program.Umask == "" {
			program.Umask = "022"
		}

		// Actualizar el mapa con los valores por defecto
		config.Programs[name] = program
	}

	return &config, nil
}
