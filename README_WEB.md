# Taskmaster Web Interface

Se ha añadido una interfaz web al proyecto Taskmaster que permite monitorizar los logs en tiempo real mediante WebSockets.

## Componentes Creados

### 1. Servidor WebSocket (`internal/web/websocket.go`)
- **Hub**: Gestor central de conexiones WebSocket
- **Client**: Maneja conexiones individuales
- **Message Types**: Estructuras para logs y status
- **Real-time Broadcasting**: Envío de logs en tiempo real a todos los clientes conectados

### 2. Servidor HTTP (`internal/web/server.go`)
- Servidor HTTP simple que sirve archivos estáticos
- Endpoint `/ws` para conexiones WebSocket
- Endpoint `/api/status` para obtener estado de procesos
- Servicio de archivos estáticos desde `web/static/`

### 3. Interfaz Web (`web/static/index.html`)
- Dashboard responsivo con diseño oscuro
- Conexión automática a WebSocket con reconexión
- Visualización de logs en tiempo real con colores por nivel
- Lista de procesos con estados
- Controles para limpiar logs y toggle de auto-scroll
- Estadísticas de procesos (total y activos)

### 4. Integración con Logger (`internal/logger/logger.go`)
- Interfaz `LogBroadcaster` para broadcasting
- Método `SetBroadcaster()` para configurar el hub
- Envío automático de logs al WebSocket cuando están disponibles

### 5. Modificaciones en Main (`cmd/taskmaster/main.go`)
- Nuevo flag `--web-port` para configurar puerto (default: 8080)
- Inicialización del servidor web en background
- Integración del hub con el logger para broadcasting

## Uso

### Iniciar Taskmaster con interfaz web:
```bash
./taskmaster --web-port=8080
```

### Acceder a la interfaz web:
```
http://localhost:8080
```

## Características

- **Logs en tiempo real**: Todos los logs del sistema se muestran instantáneamente
- **Estado de procesos**: Visualización en tiempo real del estado de todos los procesos
- **Conexión automática**: Reconexión automática si se pierde la conexión
- **Interface responsive**: Funciona en dispositivos móviles y desktop
- **Controles de logs**: Limpieza de logs y control de auto-scroll
- **Estadísticas**: Contador de procesos totales y activos

## Arquitectura

```
Taskmaster Logger → WebSocket Hub → Cliente Web
                 ↓
            Broadcast a todos los clientes conectados
```

Los logs se envían tanto al archivo como a la consola (como antes) y adicionalmente se broadcastean a todos los clientes web conectados mediante WebSocket.

## Tipos de Mensajes WebSocket

### Log Message
```json
{
  "type": "log",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "level": "INFO",
    "message": "Process started",
    "program": "taskmaster"
  }
}
```

### Status Message
```json
{
  "type": "status", 
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "program_name": [
      {
        "pid": 1234,
        "state": "RUNNING"
      }
    ]
  }
}
```

La interfaz web es completamente funcional y se integra perfectamente con el sistema existente de Taskmaster.