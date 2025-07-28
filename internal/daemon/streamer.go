package daemon

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"taskmaster/internal/process"
	"time"
)

// LogStreamer maneja el streaming de output de procesos
type LogStreamer struct {
	conn     net.Conn
	instance string
	stdout   string
	stderr   string
	cancel   context.CancelFunc
}

// NewLogStreamer crea un nuevo streamer para una instancia
func NewLogStreamer(conn net.Conn, instance, stdout, stderr string) *LogStreamer {
	return &LogStreamer{
		conn:     conn,
		instance: instance,
		stdout:   stdout,
		stderr:   stderr,
	}
}

// Start inicia el streaming de logs
func (ls *LogStreamer) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	ls.cancel = cancel

	// Verificar que los archivos de log existan
	if ls.stdout == "" && ls.stderr == "" {
		return fmt.Errorf("no log files configured for instance %s", ls.instance)
	}

	// Enviar mensaje de confirmación de attach
	attachedMsg := ServerResponse{
		Type:      RespAttached,
		Success:   true,
		Message:   fmt.Sprintf("Attached to %s. Press Ctrl+C to detach.", ls.instance),
		Timestamp: time.Now(),
	}

	encoder := json.NewEncoder(ls.conn)
	if err := encoder.Encode(attachedMsg); err != nil {
		return fmt.Errorf("failed to send attached confirmation: %w", err)
	}

	// Iniciar streaming de stdout y stderr
	if ls.stdout != "" && ls.stdout != "/dev/null" {
		go ls.streamFile(ctx, ls.stdout, "stdout")
	}

	if ls.stderr != "" && ls.stderr != "/dev/null" {
		go ls.streamFile(ctx, ls.stderr, "stderr")
	}

	return nil
}

// Stop detiene el streaming
func (ls *LogStreamer) Stop() {
	if ls.cancel != nil {
		ls.cancel()

		// Enviar mensaje de detach
		detachedMsg := ServerResponse{
			Type:      RespDetached,
			Success:   true,
			Message:   fmt.Sprintf("Detached from %s", ls.instance),
			Timestamp: time.Now(),
		}

		encoder := json.NewEncoder(ls.conn)
		encoder.Encode(detachedMsg)
	}
}

// streamFile lee un archivo de log y envía las líneas por el socket
func (ls *LogStreamer) streamFile(ctx context.Context, filename, source string) {
	// Abrir archivo
	file, err := os.Open(filename)
	if err != nil {
		// Si el archivo no existe, crearlo vacío y esperar
		file, err = os.Create(filename)
		if err != nil {
			return
		}
	}
	defer file.Close()

	// Ir al final del archivo (como tail -f)
	file.Seek(0, 2)

	scanner := bufio.NewScanner(file)
	encoder := json.NewEncoder(ls.conn)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if scanner.Scan() {
				line := scanner.Text()
				
				// Enviar línea por el socket
				streamMsg := StreamMessage{
					Type:      RespStream,
					Instance:  ls.instance,
					Content:   line,
					Source:    source,
					Timestamp: time.Now(),
				}

				if err := encoder.Encode(streamMsg); err != nil {
					return // Cliente desconectado
				}
			} else {
				// No hay más líneas, esperar un poco
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

// findProcessInstance busca una instancia específica de proceso
func (s *Server) findProcessInstance(instanceName string) (*ProcessInstance, error) {
	status := s.manager.GetStatus()
	
	for _, instances := range status {
		for _, instance := range instances {
			if instance.Name == instanceName {
				// Verificar que esté corriendo
				if instance.State != process.StateRunning {
					return nil, fmt.Errorf("process %s is not running (state: %s)", instanceName, instance.State)
				}
				return &ProcessInstance{
					Name:   instance.Name,
					Stdout: instance.Config.Stdout,
					Stderr: instance.Config.Stderr,
				}, nil
			}
		}
	}
	
	return nil, fmt.Errorf("process instance %s not found", instanceName)
}

// ProcessInstance simplificada para attach
type ProcessInstance struct {
	Name   string
	Stdout string
	Stderr string
}
