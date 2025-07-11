# **Mini-Taskmaster - Guía de Aprendizaje**

## **Conceptos Clave Demostrados**

### **1. Gestión de Procesos en Go**
```go
// Crear un proceso
cmd := exec.Command("sh", "-c", command)
cmd.Start()  // Inicia el proceso
cmd.Wait()   // Espera a que termine
```

**Conceptos importantes:**
- `exec.Command()` crea un proceso hijo
- `cmd.Start()` inicia el proceso sin bloquear
- `cmd.Wait()` espera a que termine el proceso
- `cmd.Process.Pid` obtiene el PID del proceso

### **2. Configuración YAML**
```yaml
programs:
  mi_app:
    cmd: "sleep 10"
    numprocs: 2
    autostart: true
```

**Conceptos importantes:**
- Estructura de datos clara y legible
- Validación de configuración
- Valores por defecto para campos opcionales

### **3. Concurrencia con Goroutines**
```go
// Monitorear proceso en background
go m.monitorProcess(process)
```

**Conceptos importantes:**
- `go` keyword para ejecutar en paralelo
- Goroutines para monitoreo no bloqueante
- Sincronización con `sync.RWMutex`

### **4. Señales del Sistema**
```go
// Enviar señal SIGTERM para terminar graciosamente
process.Cmd.Process.Signal(syscall.SIGTERM)

// Verificar si proceso sigue vivo
process.Cmd.Process.Signal(syscall.Signal(0))
```

**Conceptos importantes:**
- `SIGTERM` para terminación graceful
- `SIGKILL` para terminación forzada
- Señal 0 para verificar existencia

### **5. Estados de Procesos**
```go
type ProcessState int

const (
    StateStopped ProcessState = iota
    StateStarting
    StateRunning
    StateFailed
)
```

**Conceptos importantes:**
- Enum para estados bien definidos
- Transiciones de estado claras
- Método `String()` para representación

## **Flujo de Ejecución**

### **1. Inicialización**
```
main.go → config.Load() → process.NewManager() → shell.New()
```

### **2. Autostart**
```
shell.Run() → startAutoStartProcesses() → manager.StartProgram()
```

### **3. Comando Start**
```
shell.executeCommand() → startProgram() → manager.StartProgram() → monitorProcess()
```

### **4. Monitoreo**
```
monitorProcess() → cmd.Wait() → actualizar estado → log resultado
```

## **Patrones de Diseño Utilizados**

### **1. Manager Pattern**
- `process.Manager` centraliza la gestión de procesos
- Encapsula la lógica de inicio/parada/estado
- Maneja concurrencia con mutex

### **2. Factory Pattern**
- `process.NewManager()` crea instancias configuradas
- `shell.New()` crea shell con dependencias

### **3. Command Pattern**
- Shell interpreta comandos de texto
- Cada comando tiene su handler específico

## **Aspectos de Seguridad**

### **1. Validación de Entrada**
```go
if len(args) == 0 {
    fmt.Println("❌ Uso: start <nombre_programa>")
    return false
}
```

### **2. Manejo de Errores**
```go
if err := cmd.Start(); err != nil {
    return fmt.Errorf("error iniciando proceso: %w", err)
}
```

### **3. Race Conditions**
```go
m.mutex.RLock()
defer m.mutex.RUnlock()
```

## **Mejoras Posibles**

1. **Logging a Archivo**: Agregar logs persistentes
2. **Reinicio Automático**: Implementar políticas de reinicio
3. **Timeouts**: Agregar timeouts para operaciones
4. **Configuración Avanzada**: Más opciones de configuración
5. **Métricas**: Estadísticas de uso y rendimiento

## **Ejercicios Propuestos**

### **Básico**
1. Modificar el archivo de configuración
2. Agregar un nuevo programa
3. Cambiar el número de procesos

### **Intermedio**
1. Implementar comando `list` para ver programas disponibles
2. Agregar timeout para comandos stop
3. Implementar reinicio automático

### **Avanzado**
1. Agregar logging a archivo
2. Implementar señales SIGHUP para recargar config
3. Agregar métricas de CPU/memoria

## **Comandos de Prueba**

```bash
# Ejecutar el programa
go run cmd/mini-taskmaster/main.go

# En la shell:
status          # Ver estado
start counter   # Iniciar contador
stop counter    # Detener contador
restart test_app # Reiniciar test_app
exit            # Salir
```

## **Recursos Adicionales**

- [Go exec package](https://pkg.go.dev/os/exec)
- [Go syscall package](https://pkg.go.dev/syscall)
- [YAML parsing](https://pkg.go.dev/gopkg.in/yaml.v3)
- [Goroutines](https://go.dev/tour/concurrency/1)
- [Mutex](https://pkg.go.dev/sync#Mutex)