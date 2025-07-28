package process

import (
	"os"
	"strings"
)

// ProcessTeeWriter escribe a archivo Y envía al WebSocket línea por línea
type ProcessTeeWriter struct {
	file        *os.File
	broadcaster LogBroadcaster
	programName string
	source      string // "stdout" o "stderr"
	buffer      []byte
}

// NewProcessTeeWriter crea un nuevo TeeWriter
func NewProcessTeeWriter(filepath, programName, source string, broadcaster LogBroadcaster) (*ProcessTeeWriter, error) {
	// Crear directorio si no existe
	if err := os.MkdirAll(filepath[:strings.LastIndex(filepath, "/")], 0755); err != nil {
		return nil, err
	}

	// Abrir/crear archivo
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &ProcessTeeWriter{
		file:        file,
		broadcaster: broadcaster,
		programName: extractProgramBaseName(programName),
		source:      source,
		buffer:      make([]byte, 0, 1024),
	}, nil
}

// Write implementa io.Writer
func (ptw *ProcessTeeWriter) Write(data []byte) (int, error) {
	// Escribir al archivo primero
	n, err := ptw.file.Write(data)
	if err != nil {
		return n, err
	}

	// Sincronizar archivo (flush)
	ptw.file.Sync()

	// Procesar datos para WebSocket línea por línea
	ptw.processForWebSocket(data)

	return n, nil
}

// processForWebSocket procesa los datos y envía líneas completas al WebSocket
func (ptw *ProcessTeeWriter) processForWebSocket(data []byte) {
	if ptw.broadcaster == nil {
		return
	}

	// Añadir datos al buffer
	ptw.buffer = append(ptw.buffer, data...)

	// Procesar líneas completas
	for {
		newlineIndex := -1
		for i, b := range ptw.buffer {
			if b == '\n' || b == '\r' {
				newlineIndex = i
				break
			}
		}

		if newlineIndex == -1 {
			// No hay línea completa, esperar más datos
			break
		}

		// Extraer línea completa
		line := string(ptw.buffer[:newlineIndex])
		
		// Limpiar línea (remover espacios y caracteres de control)
		line = strings.TrimSpace(line)

		// Enviar al WebSocket si no está vacía
		if line != "" {
			level := ptw.determineLogLevel(line)
			ptw.broadcaster.BroadcastLog(level, line, ptw.programName)
		}

		// Remover línea procesada del buffer
		// +1 para incluir el \n, +1 más si hay \r\n
		skipBytes := newlineIndex + 1
		if skipBytes < len(ptw.buffer) && ptw.buffer[newlineIndex] == '\r' && ptw.buffer[skipBytes] == '\n' {
			skipBytes++
		}
		ptw.buffer = ptw.buffer[skipBytes:]
	}

	// Mantener buffer razonable (evitar memory leak)
	if len(ptw.buffer) > 4096 {
		// Si hay una línea muy larga, enviarla como está
		line := string(ptw.buffer)
		line = strings.TrimSpace(line)
		if line != "" {
			level := ptw.determineLogLevel(line)
			ptw.broadcaster.BroadcastLog(level, line, ptw.programName)
		}
		ptw.buffer = ptw.buffer[:0]
	}
}

// determineLogLevel determina el nivel de log basado en contenido
func (ptw *ProcessTeeWriter) determineLogLevel(line string) string {
	// Si es stderr, generalmente es error o warning
	if ptw.source == "stderr" {
		return "ERROR"
	}

	// Para stdout, analizar contenido
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "error") || strings.Contains(lower, "fail"):
		return "ERROR"
	case strings.Contains(lower, "warning") || strings.Contains(lower, "warn"):
		return "WARN"
	case strings.Contains(lower, "debug"):
		return "DEBUG"
	default:
		return "INFO"
	}
}

// Close cierra el archivo
func (ptw *ProcessTeeWriter) Close() error {
	if ptw.file != nil {
		return ptw.file.Close()
	}
	return nil
}

// extractProgramBaseName extrae el nombre base del programa (sin _0, _1, etc.)
func extractProgramBaseName(instanceName string) string {
	// Buscar el último _ seguido de números
	for i := len(instanceName) - 1; i >= 0; i-- {
		if instanceName[i] == '_' {
			// Verificar si lo que sigue son solo números
			rest := instanceName[i+1:]
			if isNumeric(rest) {
				return instanceName[:i]
			}
		}
	}
	return instanceName
}

// isNumeric verifica si una string contiene solo números
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}