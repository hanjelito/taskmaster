package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	fmt.Println("🔨 Building Taskmaster...")

	// Obtener directorio raíz del proyecto
	rootDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Si estamos en /scripts, subir un nivel
	if filepath.Base(rootDir) == "scripts" {
		rootDir = filepath.Dir(rootDir)
	}

	// Cambiar al directorio raíz
	if err := os.Chdir(rootDir); err != nil {
		fmt.Printf("Error changing to root directory: %v\n", err)
		os.Exit(1)
	}

	// Construir el proyecto
	cmd := exec.Command("go", "build", "-o", "taskmaster", "./cmd/taskmaster")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Build failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Build successful! Executable: ./taskmaster")
	fmt.Println("📋 Usage: ./taskmaster -config configs/example.yml")
}
