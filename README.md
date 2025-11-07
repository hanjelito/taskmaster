# 🚀 Taskmaster

Un demonio de control de trabajos similar a **supervisor**, implementado en Go.

## 📋 Descripción

Taskmaster es un sistema de gestión de procesos que permite:
- Iniciar, detener y reiniciar programas automáticamente
- Monitorear el estado de los procesos en tiempo real
- Recargar configuración sin interrumpir procesos no modificados
- Control interactivo mediante shell integrado
- **Interfaz web con WebSockets** para monitoreo en tiempo real
- Logging completo de eventos


# Instalar dependencias
make deps

# Compilar
make build

# Instalar en el sistema (opcional)
make install
```

## 🚀 Uso

### Inicio básico
```bash
# Con configuración por defecto
./taskmaster

# Con configuración personalizada
./taskmaster -config configs/mi-config.yml

# Con interfaz web habilitada (puerto 8080)
./taskmaster --web-port=8080

# Con configuración personalizada y web
./taskmaster -config configs/mi-config.yml --web-port=8080

# Usando Makefile
make run
make run-config CONFIG=configs/production.yml
make run-web
make run-web-config CONFIG=configs/production.yml
```

### Comandos del shell
Una vez iniciado, puedes usar estos comandos:

```
taskmaster> help
📚 Available commands:
  help     - Show this help message
  status   - Show status of all programs
  start    - Start a program (ej: start test_program)
  stop     - Stop a program (ej: stop test_program)
  restart  - Restart a program (ej: restart test_program)
  reload   - Reload configuration file
  quit/exit - Exit taskmaster
```

### Ejemplos de uso
```bash
taskmaster> status
NAME                 STATE        PID      UPTIME     RESTARTS
----------------------------------------------------------------------
test_program_0       STOPPED      -        N/A        0       
test_program_1       STOPPED      -        N/A        0       
logger_program_0     RUNNING      12767    185s       2

taskmaster> start test_program
[2025-07-22 00:58:18] INFO: Starting program test_program...
[2025-07-22 00:58:18] INFO: Process test_program_0 successfully started

taskmaster> stop test_program
[2025-07-22 00:58:25] INFO: Stopping program test_program...
[2025-07-22 00:58:25] INFO: Process test_program_0 stopped

taskmaster> restart logger_program
[2025-07-22 00:58:30] INFO: Restarting program logger_program...
[2025-07-22 00:58:30] INFO: Process logger_program_0 restarted successfully

taskmaster> reload
[2025-07-22 00:58:35] INFO: 📁 Reloading configuration from configs/example.yml...
[2025-07-22 00:58:35] INFO: ✅ Configuration reloaded successfully
```

## 📁 Estructura del proyecto

```
taskmaster/
├── cmd/taskmaster/          # Punto de entrada principal
│   └── main.go
├── internal/                # Código interno
│   ├── config/             # Gestión de configuración
│   │   ├── loader.go
│   │   └── types.go
│   ├── logger/             # Sistema de logging
│   │   └── logger.go
│   ├── process/            # Gestión de procesos
│   │   ├── manager.go
│   │   ├── instance.go
│   │   ├── lifecycle.go
│   │   ├── monitor.go
│   │   ├── reload.go
│   │   ├── helpers.go
│   │   └── types.go
│   ├── shell/              # Shell interactivo
│   │   └── shell.go
│   └── web/                # Servidor web y WebSockets
│       ├── server.go
│       └── websocket.go
├── pkg/signals/            # Utilidades de señales
│   └── signals.go
├── web/static/             # Archivos web estáticos
│   └── index.html
├── configs/                # Archivos de configuración
│   ├── example.yml
│   └── test.yml
├── .gitignore             # Archivos ignorados por git
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
# Desde otra terminal - IMPORTANTE: usar PID del proceso principal
ps aux | grep taskmaster | grep -v grep
kill -HUP <pid-taskmaster-principal>

# Ejemplo práctico:
# Si el proceso principal es PID 12099:
kill -HUP 12099
```

**⚠️ Importante**: Asegúrate de enviar la señal al **proceso principal** de taskmaster, NO a los procesos hijos gestionados. Si envías `kill -HUP` a un proceso hijo, este terminará y se reiniciará.

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
- [x] **Interfaz web con dashboard en tiempo real**
- [x] **WebSockets para actualizaciones instantáneas**
- [x] **API REST para integración externa**
- [x] Cleanup al salir

## 🌐 Interfaz Web

### Acceso a la interfaz web
```bash
# Iniciar con interfaz web
./taskmaster --web-port=8080

# Acceder desde el navegador
http://localhost:8080
```

### Funcionalidades web
- **Dashboard en tiempo real** con estado de todos los procesos
- **Logs en vivo** con WebSockets
- **Estadísticas dinámicas** (procesos activos/total)
- **API REST** disponible en `/api/status`
- **Interfaz responsive** para móviles
- **Reconexión automática** si se pierde la conexión

Para más detalles, consulta `README_WEB.md`

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

# Probar recarga de configuración con señal SIGHUP
# IMPORTANTE: Usar el PID del proceso principal taskmaster, NO de los procesos hijos
ps aux | grep taskmaster | grep -v grep  # Obtener PID del proceso principal
kill -HUP <taskmaster-main-pid>
# Ejemplo: si el proceso principal es PID 12099:
kill -HUP 12099

#para poder ver al mismo tiempo los logs cuando hacemos un kill directo al pid que cuando damos stop
# Ver logs de stdout/stderr
tail /tmp/logger.stdout

# Probar interfaz web
./taskmaster --web-port=8080
# Luego abrir http://localhost:8080 en el navegador
```

## 🔧 Comandos Makefile

```bash
make help              # Ver todos los comandos disponibles
make build             # Compilar el proyecto
make run               # Compilar y ejecutar con configuración por defecto
make run-config        # Ejecutar con configuración personalizada
make run-web           # Ejecutar con interfaz web (puerto 8080)
make run-web-config    # Ejecutar con web y configuración personalizada
make clean             # Limpiar artifacts de construcción
make deps              # Instalar dependencias (go mod download)
make test              # Ejecutar tests
make install           # Instalar en /usr/local/bin/
make uninstall         # Desinstalar del sistema
make dev               # Modo desarrollo (clean + build + create-config + run)
make create-config     # Crear archivo de configuración de ejemplo
make fmt               # Formatear código
make lint              # Ejecutar linting (requiere golangci-lint)
make check-tools       # Verificar herramientas requeridas
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