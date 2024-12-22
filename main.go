package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

// MessageRequest representa la estructura del cuerpo de la solicitud
type MessageRequest struct {
	Number  string `json:"number"`
	Message string `json:"message"`
}

// Credentials estructura para almacenar credenciales
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ScheduledMessage representa un mensaje programado
type ScheduledMessage struct {
	Number      string    `json:"number"`
	Message     string    `json:"message"`
	ScheduledAt time.Time `json:"scheduled_at"`
}

// Variables globales
var validCredentials Credentials
var scheduledMessages []ScheduledMessage

// LoadCredentialsFromJSON carga credenciales desde un archivo JSON
func LoadCredentialsFromJSON(filePath string) (Credentials, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Credentials{}, err
	}
	defer file.Close()

	var credentials Credentials
	err = json.NewDecoder(file).Decode(&credentials)
	return credentials, err
}

// SaveCredentialsToJSON guarda credenciales en un archivo JSON
func SaveCredentialsToJSON(filePath string, credentials Credentials) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(credentials)
}

// LoadScheduledMessages carga mensajes programados desde un archivo JSON
func LoadScheduledMessages(filePath string) ([]ScheduledMessage, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var messages []ScheduledMessage
	err = json.NewDecoder(file).Decode(&messages)
	return messages, err
}

// SaveScheduledMessages guarda mensajes programados en un archivo JSON
func SaveScheduledMessages(filePath string, messages []ScheduledMessage) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(messages)
}

// SendMessage maneja el envío de mensajes
func SendMessage(g *gin.Context, client *whatsmeow.Client) {
	// Verificar el encabezado de autorización
	if !validateBasicAuth(g) {
		return
	}

	var jsonBody MessageRequest

	if err := g.ShouldBindJSON(&jsonBody); err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !IsNumeric(jsonBody.Number) {
		g.JSON(http.StatusBadRequest, gin.H{"error": "not a number"})
		return
	}

	err := sendMessageWA(client, fmt.Sprintf("%v@s.whatsapp.net", jsonBody.Number), jsonBody.Message)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	g.JSON(http.StatusOK, gin.H{"response": "Sent to " + jsonBody.Number + ": " + jsonBody.Message})
}

// RecvMessage maneja la recepción de mensajes
func RecvMessage(g *gin.Context) {
	// Verificar el encabezado de autorización
	if !validateBasicAuth(g) {
		return
	}

	number := g.Query("number") // Recuperar el parámetro "number" de la consulta

	if number == "" {
		g.JSON(http.StatusBadRequest, gin.H{"error": "number parameter is required"})
		return
	}

	g.JSON(http.StatusOK, gin.H{"message": "Received number: " + number})
}

// validateBasicAuth verifica el encabezado de autorización para Basic Auth
func validateBasicAuth(g *gin.Context) bool {
	authHeader := g.GetHeader("Authorization")
	if authHeader == "" {
		g.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
		return false
	}

	// Extraer el token del encabezado
	token := strings.TrimPrefix(authHeader, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		g.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header"})
		return false
	}

	// Dividir la cadena decodificada en usuario y contraseña
	credentials := strings.Split(string(decoded), ":")
	if len(credentials) != 2 || !validateCredentials(credentials[0], credentials[1]) {
		g.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return false
	}

	return true
}

// validateCredentials verifica las credenciales proporcionadas
func validateCredentials(username, password string) bool {
	return username == validCredentials.Username && password == validCredentials.Password
}

// IsNumeric verifica si una cadena es un número válido
func IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// sendMessageWA envía un mensaje de WhatsApp
func sendMessageWA(client *whatsmeow.Client, recipient string, messageText string) error {
	message := &waProto.Message{
		Conversation: proto.String(messageText),
	}

	jid, err := types.ParseJID(recipient)
	if err != nil {
		return err
	}

	_, err = client.SendMessage(context.Background(), jid, message)
	return err
}

// eventHandler maneja eventos de WhatsApp
func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		fmt.Println("Received a message!", v.Message.GetConversation())
	}
}

// LoginWhatsapp inicia sesión en WhatsApp Web
func LoginWhatsapp() *whatsmeow.Client {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New("sqlite3", "file:filestore.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		// No ID almacenado, nuevo inicio de sesión
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Ya iniciado sesión, solo conectar
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	return client
}

// ScheduleMessagesProgrammed programa y envía mensajes programados
func ScheduleMessagesProgrammed(client *whatsmeow.Client) {
	for _, msg := range scheduledMessages {
		time.Sleep(time.Until(msg.ScheduledAt)) // Esperar hasta la hora programada
		err := sendMessageWA(client, msg.Number, msg.Message)
		if err != nil {
			log.Printf("Error sending scheduled message: %v", err)
		} else {
			log.Printf("Scheduled message sent to %s: %s", msg.Number, msg.Message)
		}
	}
}

func main() {
	port := "45981"

	// Cargar credenciales desde un archivo JSON
	credentials, err := LoadCredentialsFromJSON("credentials.json")
	if err != nil {
		panic("Failed to load credentials: " + err.Error())
	}
	validCredentials = credentials

	// Cargar mensajes programados desde un archivo JSON
	scheduledMessages, err = LoadScheduledMessages("scheduled_messages.json")
	if err != nil {
		panic("Failed to load scheduled messages: " + err.Error())
	}

	gin.SetMode(gin.ReleaseMode)
	clientWhatsapp := LoginWhatsapp()

	// Iniciar el envío de mensajes programados
	go ScheduleMessagesProgrammed(clientWhatsapp)

	r := gin.Default()
	v1 := r.Group("/api/v1")
	v1.POST("/sendMessage", func(c *gin.Context) {
		SendMessage(c, clientWhatsapp)
	})
	v1.GET("/recvMessage", RecvMessage)
	log.Println("Running on port " + port)
	r.Run(":" + port)
}