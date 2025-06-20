```markdown
# Pasos de instalación para Evolution API (WhatsApp API)

## 📋 Requisitos previos
1. Instalar **Go (Golang)**:
   - Descarga oficial: [https://go.dev/dl/](https://go.dev/dl/)
   - Verifica la instalación:  
     ```bash
     go version
     ```

2. Tener una **cuenta de WhatsApp activa** en tu teléfono móvil

## 🚀 Instalación y configuración

### 1. Clonar el repositorio
```bash
git clone https://github.com/0xC1pher/whatsapp-api.git
cd whatsapp-api
```

### 2. Instalar dependencias
```bash
go mod tidy
```

### 3. Ejecutar la API
```bash
go run main.go
```

### 4. Vincular WhatsApp (Primer uso)
1. Al iniciar, la API generará un **código QR** en la terminal o en un archivo `qr.png`
2. En tu teléfono:
   - Abre WhatsApp
   - Toca ⋮ (Menú) → **Dispositivos vinculados** → **Vincular un dispositivo**
   - Escanea el código QR mostrado por la API

### 5. Usar la API (Ejemplo de envío)
1. La API expone endpoints como `POST /send-message`
2. Ejemplo de solicitud usando `curl`:
```bash
curl -X POST http://localhost:8080/send-message \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5491112345678",
    "message": "Hola  API"
  }'
```

### 6. Automatizar envíos
Crea un script que recorra una lista de números:
```python
import requests

numeros = ["5491112345678", "5491123456789"]
api_url = "http://localhost:8080/send-message"

for numero in numeros:
    payload = {
        "phone": numero,
        "message": "Mensaje automatizado"
    }
    response = requests.post(api_url, json=payload)
    print(f"Enviado a {numero}: {response.status_code}")
```

## ⚙️ Configuración avanzada
- **Variables de entorno**: Crea un archivo `.env` para personalizar:
  ```
  PORT=3000
  SESSION_NAME=my_session
  LOG_LEVEL=debug
  ```
- **Persistencia**: Las sesiones se guardan en `./sessions/` (no eliminar esta carpeta)

## 🚨 Notas importantes
- Mantén tu **teléfono con conexión a Internet** mientras uses la API
- Si ves "dispositivo vinculado" en WhatsApp, **no pulses desconectar**
- Para reiniciar sesión: Elimina el archivo de sesión correspondiente en `./sessions/`
```

### 📌 Observaciones clave:
1. **Requisito WhatsApp**: Necesitas una cuenta de WhatsApp real para vincular
2. **Funcionamiento**: La API actúa como puente entre tu servidor y WhatsApp Web
3. **Seguridad**: Nunca compartas los archivos de sesión (`session.gob`)
4. **Alternativa oficial**: Para uso profesional considera [WhatsApp Business API](https://developers.facebook.com/docs/whatsapp/cloud-api)
