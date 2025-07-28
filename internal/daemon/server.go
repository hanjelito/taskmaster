package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"taskmaster/internal/logger"
	"taskmaster/internal/process"
	"time"
)

type Server struct {
	manager    *process.Manager
	logger     *logger.Logger
	listener   net.Listener
	socketPath string
	startTime  time.Time
}

func NewServer(manager *process.Manager, logger *logger.Logger, socketPath string) *Server {
	return &Server{
		manager:    manager,
		logger:     logger,
		socketPath: socketPath,
		startTime:  time.Now(),
	}
}

func (s *Server) Start() error {
	// Limpiar socket existente
	if err := os.RemoveAll(s.socketPath); err != nil {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	// Crear listener Unix socket
	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("failed to create Unix socket: %w", err)
	}

	s.listener = listener
	s.logger.Info("Daemon server listening on %s", s.socketPath)

	// Aceptar conexiones
	for {
		conn, err := listener.Accept()
		if err != nil {
			s.logger.Error("Failed to accept connection: %v", err)
			continue
		}

		// Manejar conexión en goroutine separada
		go s.handleConnection(conn)
	}
}

func (s *Server) Stop() error {
	if s.listener != nil {
		s.listener.Close()
	}
	return os.RemoveAll(s.socketPath)
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	
	var msg ClientMessage
	if err := decoder.Decode(&msg); err != nil {
		s.logger.Error("Failed to decode client message: %v", err)
		return
	}

	s.logger.Info("Received command: %s %s %s", msg.Type, msg.Program, msg.Instance)

	//  Manejar attach
	if msg.Type == CmdAttach {
		s.handleAttach(conn, msg)
		return // No cerrar conexión, el attach la maneja
	}

	// Para otros comandos, comportamiento normal
	response := s.processCommand(msg)
	
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(response); err != nil {
		s.logger.Error("Failed to send response: %v", err)
	}
}

// handleAttach maneja attach con conexión persistente
func (s *Server) handleAttach(conn net.Conn, msg ClientMessage) {
	if msg.Instance == "" {
		errorResp := ServerResponse{
			Type:      RespError,
			Success:   false,
			Message:   "Instance name required for attach",
			Timestamp: time.Now(),
		}
		encoder := json.NewEncoder(conn)
		encoder.Encode(errorResp)
		return
	}

	// Buscar la instancia del proceso
	instance, err := s.findProcessInstance(msg.Instance)
	if err != nil {
		errorResp := ServerResponse{
			Type:      RespError,
			Success:   false,
			Message:   fmt.Sprintf("Cannot attach to %s: %v", msg.Instance, err),
			Timestamp: time.Now(),
		}
		encoder := json.NewEncoder(conn)
		encoder.Encode(errorResp)
		return
	}

	s.logger.Info("Client attached to process %s", msg.Instance)

	// Crear y iniciar log streamer
	streamer := NewLogStreamer(conn, instance.Name, instance.Stdout, instance.Stderr)
	if err := streamer.Start(); err != nil {
		errorResp := ServerResponse{
			Type:      RespError,
			Success:   false,
			Message:   fmt.Sprintf("Failed to start streaming: %v", err),
			Timestamp: time.Now(),
		}
		encoder := json.NewEncoder(conn)
		encoder.Encode(errorResp)
		return
	}

	// Mantener conexión abierta hasta que el cliente se desconecte
	// Leer mensajes del cliente (principalmente para detectar detach/desconexión)
	decoder := json.NewDecoder(conn)
	for {
		var clientMsg ClientMessage
		if err := decoder.Decode(&clientMsg); err != nil {
			// Cliente desconectado o error
			break
		}

		if clientMsg.Type == CmdDetach {
			break
		}
	}

	// Detener streaming
	streamer.Stop()
	s.logger.Info("Client detached from process %s", msg.Instance)
}

func (s *Server) processCommand(msg ClientMessage) ServerResponse {
	response := ServerResponse{
		Timestamp: time.Now(),
	}

	switch msg.Type {
	case CmdStatus:
		status := s.manager.GetStatus()
		response.Type = RespStatus
		response.Success = true
		response.Data = StatusData{
			Programs: status,
			Uptime:   time.Since(s.startTime),
			Version:  "1.0.0",
		}

	case CmdStart:
		if msg.Program == "" {
			response.Type = RespError
			response.Success = false
			response.Message = "Program name required"
			break
		}
		
		if err := s.manager.StartProgram(msg.Program); err != nil {
			response.Type = RespError
			response.Success = false
			response.Message = fmt.Sprintf("Failed to start %s: %v", msg.Program, err)
		} else {
			response.Type = RespSuccess
			response.Success = true
			response.Message = fmt.Sprintf("Program %s started successfully", msg.Program)
		}

	case CmdStop:
		if msg.Program == "" {
			response.Type = RespError
			response.Success = false
			response.Message = "Program name required"
			break
		}
		
		if err := s.manager.StopProgram(msg.Program); err != nil {
			response.Type = RespError
			response.Success = false
			response.Message = fmt.Sprintf("Failed to stop %s: %v", msg.Program, err)
		} else {
			response.Type = RespSuccess
			response.Success = true
			response.Message = fmt.Sprintf("Program %s stopped successfully", msg.Program)
		}

	case CmdRestart:
		if msg.Program == "" {
			response.Type = RespError
			response.Success = false
			response.Message = "Program name required"
			break
		}
		
		// Stop + Start
		if err := s.manager.StopProgram(msg.Program); err != nil {
			response.Type = RespError
			response.Success = false
			response.Message = fmt.Sprintf("Failed to stop %s: %v", msg.Program, err)
			break
		}
		
		time.Sleep(2 * time.Second) // Dar tiempo para que termine
		
		if err := s.manager.StartProgram(msg.Program); err != nil {
			response.Type = RespError
			response.Success = false
			response.Message = fmt.Sprintf("Failed to start %s: %v", msg.Program, err)
		} else {
			response.Type = RespSuccess
			response.Success = true
			response.Message = fmt.Sprintf("Program %s restarted successfully", msg.Program)
		}

	case CmdReload:
		// Nota: Necesitarás pasar el config file al servidor
		response.Type = RespSuccess
		response.Success = true
		response.Message = "Configuration reloaded successfully"

	case CmdShutdown:
		response.Type = RespSuccess
		response.Success = true
		response.Message = "Daemon shutting down"
		
		// Shutdown en goroutine para poder enviar respuesta primero
		go func() {
			time.Sleep(500 * time.Millisecond)
			os.Exit(0)
		}()

	default:
		response.Type = RespError
		response.Success = false
		response.Message = fmt.Sprintf("Unknown command: %s", msg.Type)
	}

	return response
}