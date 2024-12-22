# 🚀 WhatsApp Sender API

Este proyecto es una API REST que permite enviar y recibir mensajes de WhatsApp utilizando la biblioteca [whatsmeow](https://github.com/tulir/whatsmeow).

## 🌟 Características

- **Envío de mensajes**: Envía mensajes de texto a números de WhatsApp específicos.
- **Recepción de mensajes**: Simula la recepción de mensajes para pruebas.
- **Autenticación básica**: Protege las rutas con autenticación básica (username y password).
- **Integración con WhatsApp Web**: Inicia sesión en WhatsApp Web mediante un código QR.

## 🛠️ Tecnologías utilizadas

- **Lenguaje**: Go (Golang)
- **Biblioteca de WhatsApp**: [whatsmeow](https://github.com/tulir/whatsmeow)
- **Framework web**: [Gin](https://github.com/gin-gonic/gin)
- **Base de datos**: SQLite (para almacenar datos de sesión)

## 📝 Instrucciones

1. **Configura las credenciales**:
   - Crea un archivo `credentials.txt` con el formato `username:password`.

2. **Ejecuta el servidor**:
   - Ejecuta el programa con `go run main.go`.

3. **Envía mensajes**:
   - Usa la ruta `/api/v1/sendMessage` para enviar mensajes.

4. **Recibe mensajes**:
   - Usa la ruta `/api/v1/recvMessage` para simular la recepción de mensajes.

## 📄 Documentación

- **Autenticación**: Usa el encabezado `Authorization: Basic base64(username:password)`.
- **Ejemplo de solicitud**:
  ```json
  {
    "number": "1234567890",
    "message": "Hola, este es un mensaje de prueba"
  }
