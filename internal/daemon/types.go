package daemon

import (
	"taskmaster/internal/process"
	"time"
)

// MessageType define los tipos de mensajes entre cliente y servidor
type MessageType string

const (
	// Comandos del cliente al servidor
	CmdStatus    MessageType = "status"
	CmdStart     MessageType = "start"
	CmdStop      MessageType = "stop"
	CmdRestart   MessageType = "restart"
	CmdReload    MessageType = "reload"
	CmdShutdown  MessageType = "shutdown"
	CmdAttach    MessageType = "attach"
	CmdDetach    MessageType = "detach"
	
	// Respuestas del servidor al cliente
	RespSuccess  MessageType = "success"
	RespError    MessageType = "error"
	RespStatus   MessageType = "status_data"
	RespStream   MessageType = "stream"
	RespAttached MessageType = "attached"
	RespDetached MessageType = "detached"
)

// ClientMessage representa un mensaje del cliente al servidor
type ClientMessage struct {
	Type      MessageType `json:"type"`
	Program   string      `json:"program,omitempty"`   // Para start/stop/restart
	Instance  string      `json:"instance,omitempty"`  // Para attach (nombre específico de instancia)
	Timestamp time.Time   `json:"timestamp"`
}

// ServerResponse representa una respuesta del servidor al cliente
type ServerResponse struct {
	Type      MessageType `json:"type"`
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`  // Para status
	Timestamp time.Time   `json:"timestamp"`
}

// StatusData contiene información de estado de procesos
type StatusData struct {
	Programs map[string][]*process.ProcessInstance `json:"programs"`
	Uptime   time.Duration                         `json:"uptime"`
	Version  string                                `json:"version"`
}

// StreamMessage para envío de output en tiempo real
type StreamMessage struct {
	Type      MessageType `json:"type"`
	Instance  string      `json:"instance"`
	Content   string      `json:"content"`    // Línea de output del proceso
	Source    string      `json:"source"`     // "stdout" o "stderr"
	Timestamp time.Time   `json:"timestamp"`
}