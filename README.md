# 🚀 Taskmaster

Un demonio de control de trabajos similar a **supervisor**, implementado en Go.

## 📋 Descripción

Taskmaster es un sistema de gestión de procesos que permite:
- Iniciar, detener y reiniciar programas automáticamente
- Monitorear el estado de los procesos
- Recargar configuración sin interrumpir procesos no modificados
- Control interactivo mediante shell integrado
- Logging completo de eventos

## 🛠️ Instalación

### Prerrequisitos
- Go 1.21 o superior
- Sistema Unix/Linux (probado en Ubuntu, macOS)

### Compilación
```bash
# Clonar el repositorio
git clone <repository-url>
cd taskmaster

# Instalar dependencias
make deps

# Compilar
make build

# O usar el script de construcción
make build-script
```

## 🚀 Uso

### Inicio básico
```bash
# Con configuración por defecto
./taskmaster

# Con configuración personalizada
./taskmaster -config configs/mi-config.yml

# Usando Makefile
make run
make run-config CONFIG=configs/production.yml
```

### Comandos del shell
Una vez iniciado, puedes usar estos comandos:

```
taskmaster> help
📚 Available commands:
  help     - Show this help message
  status   - Show status of all programs
  start    - Start a program
  stop     - Stop a program
  restart  - Restart a program
  reload   - Reload configuration file
  quit/exit - Exit taskmaster
```

### Ejemplos de uso
```bash
taskmaster> status
taskmaster> start test_program
taskmaster> stop test_program
taskmaster> restart logger_program
taskmaster> reload
```

## 📁 Estructura del proyecto

```
taskmaster/
├── cmd/taskmaster/          # Punto de entrada principal
│   └── main.go
├── internal/                # Código interno
│   ├── config/             # Gestión de configuración
│   │   └── config.go
│   ├── logger/             # Sistema de logging
│   │   └── logger.go
│   ├── process/            # Gestión de procesos
│   │   └── manager.go
│   └── shell/              # Shell interactivo
│       └── shell.go
├── pkg/signals/            # Utilidades de señales
│   └── signals.go
├── configs/                # Archivos de configuración
│   ├── example.yml
│   └── test.yml
├── scripts/                # Scripts auxiliares
│   └── build.go
├── Makefile               # Comandos de construcción
├── go.mod                 # Dependencias Go
└── README.md
```

## ⚙️ Configuración

### Formato del archivo de configuración (YAML)

```yaml
programs:
  mi_programa:
    cmd: "mi-comando --args"           # Comando a ejecutar
    numprocs: 2                       # Número de procesos
    autostart: true                   # Iniciar automáticamente
    autorestart: unexpected           # always, never, unexpected
    exitcodes: [0, 2]                # Códigos de salida esperados
    starttime: 3                      # Tiempo para considerar "iniciado exitosamente"
    startretries: 3                   # Reintentos antes de abandonar
    stopsignal: TERM                  # Señal para detener gracefully
    stoptime: 10                      # Tiempo antes de KILL
    stdout: /tmp/programa.stdout      # Redirección stdout
    stderr: /tmp/programa.stderr      # Redirección stderr
    env:                              # Variables de entorno
      MI_VAR: "valor"
    workingdir: /tmp                  # Directorio de trabajo
    umask: "022"                      # Umask del proceso
```

### Opciones de configuración

| Campo | Descripción | Valores | Por defecto |
|-------|-------------|---------|-------------|
| `cmd` | Comando a ejecutar | string | **requerido** |
| `numprocs` | Número de procesos | int | 1 |
| `autostart` | Iniciar automáticamente | bool | false |
| `autorestart` | Política de reinicio | always/never/unexpected | unexpected |
| `exitcodes` | Códigos de salida esperados | []int | [0] |
| `starttime` | Tiempo considerado iniciado | int (segundos) | 1 |
| `startretries` | Intentos de reinicio | int | 3 |
| `stopsignal` | Señal de parada | TERM/KILL/INT/USR1/USR2 | TERM |
| `stoptime` | Timeout antes de KILL | int (segundos) | 10 |
| `stdout` | Redirección stdout | path o /dev/null | - |
| `stderr` | Redirección stderr | path o /dev/null | - |
| `env` | Variables de entorno | map[string]string | - |
| `workingdir` | Directorio de trabajo | path | - |
| `umask` | Umask del proceso | string octal | 022 |

## 🔄 Recarga de configuración

### Mediante comando shell
```bash
taskmaster> reload
```

### Mediante señal SIGHUP
```bash
# Desde otra terminal
kill -HUP <pid-taskmaster>
```

**Comportamiento de recarga:**
- ✅ Programas nuevos se inician si `autostart: true`
- ✅ Programas modificados se reinician
- ✅ Programas eliminados se detienen
- ✅ Programas sin cambios **NO** se reinician

## 📊 Estados de procesos

| Estado | Descripción |
|--------|-------------|
| `STOPPED` | Proceso detenido |
| `STARTING` | Proceso iniciándose |
| `RUNNING` | Proceso ejecutándose normalmente |
| `FAILED` | Proceso falló después de todos los reintentos |
| `RESTARTING` | Proceso reiniciándose |

## 🎯 Características implementadas

### ✅ Características básicas
- [x] Shell de control interactivo
- [x] Carga de configuración desde archivo YAML
- [x] Sistema de logging a archivo
- [x] Recarga de configuración (SIGHUP + comando)
- [x] Inicio/parada/reinicio de programas

### ✅ Opciones de configuración
- [x] Comando de ejecución
- [x] Número de procesos
- [x] Inicio automático
- [x] Política de reinicio (always/never/unexpected)
- [x] Códigos de salida esperados
- [x] Tiempo de inicio exitoso
- [x] Intentos de reinicio
- [x] Señal de parada
- [x] Timeout de parada
- [x] Redirección stdout/stderr
- [x] Variables de entorno
- [x] Directorio de trabajo
- [x] Umask

### ✅ Características avanzadas
- [x] Monitoreo automático de procesos
- [x] Reinicio automático según configuración
- [x] Manejo graceful de señales
- [x] Estados de proceso detallados
- [x] Logging de eventos completo
- [x] Cleanup al salir

## 🧪 Pruebas

### Crear configuración de prueba
```bash
make create-config
```

### Ejecutar en modo desarrollo
```bash
make dev
```

### Comandos útiles para probar
```bash
# Verificar procesos en el sistema
ps aux | grep sleep

# Matar proceso manualmente para probar reinicio
kill <pid>

# Monitorear logs
tail -f taskmaster.log

# Probar recarga de configuración
kill -HUP <taskmaster-pid>
```

## 🔧 Comandos Makefile

```bash
make help          # Ver todos los comandos disponibles
make build         # Compilar
make run           # Compilar y ejecutar
make clean         # Limpiar artifacts
make deps          # Instalar dependencias
make test          # Ejecutar tests
make install       # Instalar en sistema
make dev           # Modo desarrollo
```

## 📝 Logging

Los logs se guardan en `taskmaster.log` e incluyen:
- Inicio/parada de procesos
- Cambios de estado
- Reinicios automáticos
- Recarga de configuración
- Errores y eventos importantes

## 🚨 Manejo de errores

- Procesos que fallan repetidamente se marcan como `FAILED`
- Reintentos limitados según `startretries`
- Timeout configurable para parada graceful
- Logging detallado de todos los errores

## 🤝 Contribución

1. Fork el proyecto
2. Crear branch feature (`git checkout -b feature/AmazingFeature`)
3. Commit cambios (`git commit -m 'Add AmazingFeature'`)
4. Push al branch (`git push origin feature/AmazingFeature`)
5. Abrir Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT - ver el archivo [LICENSE](LICENSE) para detalles.

## 🙏 Agradecimientos

- Inspirado en [Supervisor](http://supervisord.org/)
- Proyecto educativo de [42 School](https://42.fr/)

---

**¡Disfruta usando Taskmaster! 🚀**