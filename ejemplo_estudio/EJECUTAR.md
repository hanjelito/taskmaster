# **Cómo Ejecutar Mini-Taskmaster**

## **Requisitos**
- Go 1.21 o superior
- Sistema Unix/Linux/macOS

## **Instalación**
```bash
# Entrar al directorio
cd ejemplo_estudio

# Descargar dependencias
go mod tidy

# Ejecutar el programa
go run cmd/mini-taskmaster/main.go
```

## **Uso Básico**

### **1. Iniciar la aplicación**
```bash
go run cmd/mini-taskmaster/main.go
```

Verás algo como:
```
🚀 Iniciando Mini-Taskmaster...
📁 Cargando configuración desde: configs/example.yml
✅ Configuración cargada: 3 programas
🎮 Mini-Taskmaster Shell
Escribe 'help' para ver comandos disponibles
🔄 Iniciando procesos con autostart...
🚀 Iniciando programa 'test_app' con 2 procesos
✅ Proceso test_app_0 iniciado (PID: 12345)
✅ Proceso test_app_1 iniciado (PID: 12346)
mini-taskmaster> 
```

### **2. Ver estado de procesos**
```bash
mini-taskmaster> status
```

### **3. Iniciar un programa**
```bash
mini-taskmaster> start counter
```

### **4. Detener un programa**
```bash
mini-taskmaster> stop counter
```

### **5. Reiniciar un programa**
```bash
mini-taskmaster> restart test_app
```

### **6. Ver ayuda**
```bash
mini-taskmaster> help
```

### **7. Salir**
```bash
mini-taskmaster> exit
```

## **Configuración Personalizada**

### **1. Usar archivo de configuración diferente**
```bash
go run cmd/mini-taskmaster/main.go -config mi_config.yml
```

### **2. Crear tu propio archivo de configuración**
```yaml
programs:
  mi_programa:
    cmd: "echo 'Hola mundo'; sleep 5"
    numprocs: 1
    autostart: true
    autorestart: false
```

## **Ejemplos de Prueba**

### **1. Programa que cuenta**
```bash
mini-taskmaster> start counter
mini-taskmaster> status
# Observa cómo cuenta y termina
mini-taskmaster> status
```

### **2. Programa de larga duración**
```bash
mini-taskmaster> start long_task
mini-taskmaster> status
# Espera un poco
mini-taskmaster> stop long_task
mini-taskmaster> status
```

### **3. Reiniciar programa**
```bash
mini-taskmaster> start counter
mini-taskmaster> restart counter
# El contador se reinicia desde 1
```

## **Debugging**

### **1. Ver qué procesos están corriendo**
```bash
# En otro terminal
ps aux | grep sleep
```

### **2. Matar proceso manualmente**
```bash
# En otro terminal
kill <PID>
# Luego en mini-taskmaster
mini-taskmaster> status
```

## **Logs y Salida**

- Los logs del programa aparecen en la consola
- Los procesos iniciados pueden generar su propia salida
- El programa maneja la limpieza automática al salir

## **Troubleshooting**

### **Error: "archivo no encontrado"**
```bash
# Asegúrate de estar en el directorio correcto
cd ejemplo_estudio
```

### **Error: "comando no encontrado"**
```bash
# Verificar que tienes bash instalado
which bash
```

### **Procesos no terminan**
```bash
# Usar exit para terminar limpiamente
mini-taskmaster> exit
```