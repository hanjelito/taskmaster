package main

import (
	"flag"
	"fmt"
	"log"
	"mini-taskmaster/internal/config"
	"mini-taskmaster/internal/process"
	"mini-taskmaster/internal/shell"
)

func main() {
	// Parsear argumentos de línea de comandos
	configFile := flag.String("config", "configs/example.yml", "Archivo de configuración")
	flag.Parse()

	fmt.Println("🚀 Iniciando Mini-Taskmaster...")
	
	// Cargar configuración
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("❌ Error cargando configuración: %v", err)
	}

	// Crear gestor de procesos
	manager := process.NewManager()
	
	// Crear y ejecutar shell
	sh := shell.New(manager, cfg)
	sh.Run()

	fmt.Println("✅ Mini-Taskmaster terminado")
}