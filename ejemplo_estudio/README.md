# **Mini-Taskmaster - Proyecto de Estudio**

## **Descripción**
Este es un proyecto educativo simplificado que demuestra los conceptos fundamentales de taskmaster:
- Gestión de procesos
- Configuración YAML
- Shell interactiva
- Monitoreo de procesos
- Logging básico

## **Características Implementadas**
- **Configuración YAML**: Definición de procesos en archivo config
- **Gestión de Procesos**: Iniciar, parar, reiniciar procesos
- **Shell Interactiva**: Comandos básicos (status, start, stop, restart, exit)
- **Monitoreo**: Verificación de estado de procesos
- **Logging**: Registro de eventos básico en consola

## **Estructura del Proyecto**
```
mini-taskmaster/
├── cmd/mini-taskmaster/    # Aplicación principal
├── internal/
│   ├── config/            # Configuración YAML
│   ├── process/           # Gestión de procesos
│   └── shell/             # Shell interactiva
├── configs/               # Archivos de configuración
└── go.mod                 # Dependencias
```

## **Uso**
```bash
# Ejecutar el mini-taskmaster
go run cmd/mini-taskmaster/main.go

# Comandos disponibles en el shell:
status          # Ver estado de procesos
start <nombre>  # Iniciar proceso
stop <nombre>   # Parar proceso
restart <nombre># Reiniciar proceso
exit           # Salir
```

## **Archivo de Configuración**
```yaml
programs:
  test_app:
    cmd: "sleep 10"
    numprocs: 2
    autostart: true
    autorestart: true
```

## **Objetivo Educativo**
Este proyecto te ayuda a entender:
1. **Cómo funciona la gestión de procesos en Go**
2. **Cómo cargar configuración desde YAML**
3. **Cómo crear una shell interactiva simple**
4. **Cómo monitorear procesos del sistema**
5. **Conceptos básicos de concurrencia**