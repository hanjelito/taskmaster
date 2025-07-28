package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"taskmaster/internal/process"
	"time"
)

type Client struct {
	socketPath string
}

func NewClient(socketPath string) *Client {
	return &Client{
		socketPath: socketPath,
	}
}


func (c *Client) sendCommand(cmdType MessageType, program string) (*ServerResponse, error) {
	// Conectar al daemon
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	// Enviar mensaje
	msg := ClientMessage{
		Type:      cmdType,
		Program:   program,
		Timestamp: time.Now(),
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Recibir respuesta
	decoder := json.NewDecoder(conn)
	var response ServerResponse
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	return &response, nil
}
func (c *Client) AttachToProcess(instanceName string) error {
	// Conectar al daemon
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	// Enviar comando attach
	msg := ClientMessage{
		Type:      CmdAttach,
		Instance:  instanceName,
		Timestamp: time.Now(),
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		return fmt.Errorf("failed to send attach command: %w", err)
	}

	// Manejar streaming
	return c.handleAttachSession(conn)
}

// handleAttachSession maneja la sesión de attach
func (c *Client) handleAttachSession(conn net.Conn) error {
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	// Canal para manejar Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Goroutine para manejar Ctrl+C
	go func() {
		<-sigChan
		// Enviar comando detach
		detachMsg := ClientMessage{
			Type:      CmdDetach,
			Timestamp: time.Now(),
		}
		encoder.Encode(detachMsg)
	}()

	// Leer mensajes del servidor
	for {
		var response interface{}
		if err := decoder.Decode(&response); err != nil {
			// Conexión cerrada
			break
		}

		// Determinar tipo de mensaje
		msgMap, ok := response.(map[string]interface{})
		if !ok {
			continue
		}

		msgType, ok := msgMap["type"].(string)
		if !ok {
			continue
		}

		switch MessageType(msgType) {
		case RespAttached:
			// Mensaje de confirmación de attach
			if msg, ok := msgMap["message"].(string); ok {
				fmt.Printf("\033[33m[ATTACHED]\033[0m %s\n", msg)
			}

		case RespStream:
			// Línea de output del proceso
			if content, ok := msgMap["content"].(string); ok {
				if source, ok := msgMap["source"].(string); ok {
					// Colorear según la fuente
					if source == "stderr" {
						fmt.Printf("\033[31m%s\033[0m\n", content) // Rojo para stderr
					} else {
						fmt.Printf("%s\n", content) // Normal para stdout
					}
				} else {
					fmt.Printf("%s\n", content)
				}
			}

		case RespDetached:
			// Mensaje de confirmación de detach
			if msg, ok := msgMap["message"].(string); ok {
				fmt.Printf("\033[33m[DETACHED]\033[0m %s\n", msg)
			}
			return nil

		case RespError:
			// Error del servidor
			if msg, ok := msgMap["message"].(string); ok {
				fmt.Printf("\033[31m[ERROR]\033[0m %s\n", msg)
			}
			return fmt.Errorf("attach failed")

		default:
			// Otros tipos de mensaje
			if msg, ok := msgMap["message"].(string); ok {
				fmt.Printf("[%s] %s\n", msgType, msg)
			}
		}
	}

	return nil
}

func (c *Client) GetStatus() (map[string][]*process.ProcessInstance, error) {
	resp, err := c.sendCommand(CmdStatus, "")
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("status command failed: %s", resp.Message)
	}

	// Convertir Data a StatusData
	statusData, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid status data format")
	}

	programs, ok := statusData["programs"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid programs data format")
	}

	// Convertir a formato esperado
	result := make(map[string][]*process.ProcessInstance)
	for name, instancesData := range programs {
		instances, ok := instancesData.([]interface{})
		if !ok {
			continue
		}

		var processInstances []*process.ProcessInstance
		for _, instanceData := range instances {
			instanceMap, ok := instanceData.(map[string]interface{})
			if !ok {
				continue
			}

			instance := &process.ProcessInstance{
				Name: instanceMap["name"].(string),
				PID:  int(instanceMap["pid"].(float64)),
				// ... mapear otros campos según necesidad
			}
			processInstances = append(processInstances, instance)
		}
		result[name] = processInstances
	}

	return result, nil
}

func (c *Client) StartProgram(name string) error {
	resp, err := c.sendCommand(CmdStart, name)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func (c *Client) StopProgram(name string) error {
	resp, err := c.sendCommand(CmdStop, name)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func (c *Client) RestartProgram(name string) error {
	resp, err := c.sendCommand(CmdRestart, name)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func (c *Client) ReloadConfig() error {
	resp, err := c.sendCommand(CmdReload, "")
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func (c *Client) Shutdown() error {
	resp, err := c.sendCommand(CmdShutdown, "")
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}