package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config representa la configuración general del sistema
type Config struct {
	Programs map[string]*Program `yaml:"programs"`
}

// Program representa la configuración de un programa específico
type Program struct {
	Cmd         string `yaml:"cmd"`         // Comando a ejecutar
	NumProcs    int    `yaml:"numprocs"`    // Número de procesos
	AutoStart   bool   `yaml:"autostart"`   // Iniciar automáticamente
	AutoRestart bool   `yaml:"autorestart"` // Reiniciar automáticamente
}

// Load carga la configuración desde un archivo YAML
func Load(filename string) (*Config, error) {
	fmt.Printf("📁 Cargando configuración desde: %s\n", filename)
	
	// Leer archivo
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error leyendo archivo: %w", err)
	}

	// Parsear YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parseando YAML: %w", err)
	}

	// Validar configuración
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("configuración inválida: %w", err)
	}

	fmt.Printf("✅ Configuración cargada: %d programas\n", len(config.Programs))
	return &config, nil
}

// validate verifica que la configuración sea válida
func (c *Config) validate() error {
	if len(c.Programs) == 0 {
		return fmt.Errorf("no se encontraron programas en la configuración")
	}

	for name, program := range c.Programs {
		if program.Cmd == "" {
			return fmt.Errorf("programa '%s': comando vacío", name)
		}
		if program.NumProcs <= 0 {
			program.NumProcs = 1 // Valor por defecto
		}
	}

	return nil
}