# **Taskmaster - Verificación de Requisitos**

## **FUNCIONALIDADES DEL PROGRAMA PRINCIPAL**

### **Gestión de Procesos** 
- **Estado**: `IMPLEMENTADO`
- **Funciones**: `StartAutoStartProcesses()`, `startProcessInstance()`
- **Descripción**: Inicia y mantiene procesos hijos

### **Monitoreo de Estado**
- **Estado**: `IMPLEMENTADO`
- **Funciones**: `StartPeriodicStatusCheck()`
- **Descripción**: Verifica cada 5 segundos si los procesos siguen vivos

### **Reinicio Automático**
- **Estado**: `IMPLEMENTADO`
- **Funciones**: `attemptRestart()`
- **Descripción**: Reinicia según política configurada (`always`, `never`, `unexpected`)

### **Configuración YAML**
- **Estado**: `IMPLEMENTADO`
- **Funciones**: `config.Load()`
- **Descripción**: Lee archivos de configuración .yml

### **Recarga de Configuración**
- **Estado**: `IMPLEMENTADO`
- **Funciones**: `handleSignals()`
- **Descripción**: Recarga configuración con señal SIGHUP sin parar procesos

### **Sistema de Logging**
- **Estado**: `IMPLEMENTADO`
- **Funciones**: `logger.New()`
- **Descripción**: Registra eventos en `taskmaster.log`

### **Shell Interactiva**
- **Estado**: `IMPLEMENTADO`
- **Funciones**: `shell.Run()`
- **Descripción**: Proporciona shell de control en primer plano

### **Compatibilidad VM**
- **Estado**: `IMPLEMENTADO`
- **Descripción**: Código estándar sin dependencias específicas del sistema

---

## **COMANDOS DEL SHELL DE CONTROL**

### **Comando `status`**
- **Estado**: `IMPLEMENTADO`
- **Funciones**: `showStatus()`
- **Descripción**: Muestra estado de todos los programas configurados

### **Comandos `start/stop/restart`**
- **Estado**: `IMPLEMENTADO`
- **Funciones**: `startProgram()`, `stopProgram()`, `restartProgram()`
- **Descripción**: Inicia, detiene y reinicia programas específicos

### **Comando `reload`**
- **Estado**: `IMPLEMENTADO`
- **Funciones**: `reloadConfig()`
- **Descripción**: Recarga archivo de configuración sin detener programa principal

### **Comando `quit/exit`**
- **Estado**: `IMPLEMENTADO`
- **Descripción**: Termina el programa principal de forma ordenada

---

## **CONFIGURACIONES POR PROGRAMA**

### **Comando de Lanzamiento**
- **Parámetro**: `cmd`
- **Estado**: `IMPLEMENTADO`
- **Descripción**: Comando para ejecutar el programa

### **Número de Procesos**
- **Parámetro**: `numprocs`
- **Estado**: `IMPLEMENTADO`
- **Descripción**: Cantidad de instancias del proceso a mantener

### **Autostart**
- **Parámetro**: `autostart`
- **Estado**: `IMPLEMENTADO`
- **Descripción**: Inicia automáticamente al arranque del sistema

### **Política de Reinicio**
- **Parámetro**: `autorestart`
- **Estado**: `IMPLEMENTADO`
- **Valores**: `always`, `never`, `unexpected`
- **Descripción**: Define cuándo reiniciar el proceso

### **Códigos de Salida Esperados**
- **Parámetro**: `exitcodes`
- **Estado**: `IMPLEMENTADO`
- **Descripción**: Lista de códigos de retorno considerados exitosos

### **Tiempo de Inicio Exitoso**
- **Parámetro**: `starttime`
- **Estado**: `IMPLEMENTADO`
- **Descripción**: Tiempo que debe ejecutarse para considerarse iniciado exitosamente

### **Intentos de Reinicio**
- **Parámetro**: `startretries`
- **Estado**: `IMPLEMENTADO`
- **Descripción**: Número máximo de intentos de reinicio antes de abortar

### **Señal de Parada**
- **Parámetro**: `stopsignal`
- **Estado**: `IMPLEMENTADO`
- **Descripción**: Señal para terminar graciosamente el programa

### **Tiempo de Espera**
- **Parámetro**: `stoptime`
- **Estado**: `IMPLEMENTADO`
- **Descripción**: Timeout antes de enviar KILL después de parada graceful

### **Redirección de Salida**
- **Parámetros**: `stdout`, `stderr`
- **Estado**: `IMPLEMENTADO`
- **Descripción**: Redirección de salida estándar y error a archivos

### **Variables de Entorno**
- **Parámetro**: `env`
- **Estado**: `IMPLEMENTADO`
- **Descripción**: Variables de entorno personalizadas para el proceso

### **Directorio de Trabajo**
- **Parámetro**: `workingdir`
- **Estado**: `IMPLEMENTADO`
- **Descripción**: Directorio de trabajo para la ejecución del proceso

### **Umask**
- **Parámetro**: `umask`
- **Estado**: `IMPLEMENTADO`
- **Descripción**: Permisos de archivos creados por el proceso

---

## **RESTRICCIONES TÉCNICAS**

### **Sin Privilegios Root**
- **Estado**: `CUMPLIDO`
- **Descripción**: El programa ejecuta con usuario normal

### **No Daemon Obligatorio**
- **Estado**: `CUMPLIDO`
- **Descripción**: Se puede iniciar desde shell normal

### **Limitaciones de Librerías**
- **Estado**: `CUMPLIDO`
- **Librerías Externas Utilizadas**:
  - `gopkg.in/yaml.v3` - Parsing de configuración
  - `github.com/chzyer/readline` - Shell interactiva
  - `github.com/gorilla/websocket` - Comunicación cliente/servidor (bonus)

---

## **FUNCIONALIDADES BONUS**

### **Interfaz Web**
- **Estado**: `IMPLEMENTADO`
- **Tecnología**: WebSocket en tiempo real
- **Funciones**: Monitoreo de logs y estado de procesos

### **Monitoreo Externo**
- **Estado**: `IMPLEMENTADO`
- **Descripción**: Detecta procesos matados desde terminal externo

---

## **RESUMEN**

**ESTADO GENERAL**: `COMPLETAMENTE IMPLEMENTADO`

**Funcionalidades Principales**: `8/8 IMPLEMENTADAS`

**Comandos de Shell**: `4/4 IMPLEMENTADOS`

**Configuraciones**: `12/12 IMPLEMENTADAS`

**Restricciones Técnicas**: `3/3 CUMPLIDAS`

**Funcionalidades Bonus**: `2/2 IMPLEMENTADAS`

**Total**: `100% COMPLETO`