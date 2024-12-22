package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

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

// MessageRequest represents the structure of the request body
type MessageRequest struct {
	Number  string `json:"number"`
	Message string `json:"message"`
}

// Credentials structure to hold username and password
type Credentials struct {
	Username string
	Password string
}

// Global variable to hold credentials
var validCredentials Credentials

// LoadCredentials reads the username and password from a file
func LoadCredentials(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Assuming credentials are in "username:password" format
	parts := strings.TrimSpace(string(data))
	credentialParts := strings.Split(parts, ":")
	if len(credentialParts) != 2 {
		return fmt.Errorf("invalid credentials format in file")
	}

	validCredentials = Credentials{
		Username: credentialParts[0],
		Password: credentialParts[1],
	}

	return nil
}

// SendMessage handles sending a message
func SendMessage(g *gin.Context, client *whatsmeow.Client) {
	// Check Authorization header
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

// RecvMessage handles receiving a message
func RecvMessage(g *gin.Context) {
	// Check Authorization header
	if !validateBasicAuth(g) {
		return
	}

	number := g.Query("number") // Retrieve the "number" query parameter

	if number == "" {
		g.JSON(http.StatusBadRequest, gin.H{"error": "number parameter is required"})
		return
	}

	g.JSON(http.StatusOK, gin.H{"message": "Received number: " + number})
}

// validateBasicAuth checks the Authorization header for Basic Auth
func validateBasicAuth(g *gin.Context) bool {
	authHeader := g.GetHeader("Authorization")
	if authHeader == "" {
		g.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
		return false
	}

	// Extract the token from the header
	token := strings.TrimPrefix(authHeader, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		g.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header"})
		return false
	}

	// Split the decoded string into username and password
	credentials := strings.Split(string(decoded), ":")
	if len(credentials) != 2 || !validateCredentials(credentials[0], credentials[1]) {
		g.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return false
	}

	return true
}

// validateCredentials checks the provided username and password against the loaded credentials
func validateCredentials(username, password string) bool {
	return username == validCredentials.Username && password == validCredentials.Password
}

// IsNumeric checks if a string is a valid number
func IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// sendMessageWA sends a message via WhatsApp
func sendMessageWA(client *whatsmeow.Client, recipient string, messageText string) error {
	// Create a message
	message := &waProto.Message{
		Conversation: proto.String(messageText),
	}

	log.Println("Sending to " + recipient + " with message: " + messageText)
	// Convert the recipient number into the correct WhatsApp format (JID)
	jid, err := types.ParseJID(recipient)
	if err != nil {
		return err
	}

	// Send the message
	_, err = client.SendMessage(context.Background(), jid, message)
	if err != nil {
		return err
	}

	return nil
}

// eventHandler handles WhatsApp events
func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		fmt.Println("Received a message!", v.Message.GetConversation())
	}
}

// LoginWhatsapp logs into WhatsApp Web
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
		// No ID stored, new login
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
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	return client
}

func main() {
	port := "45981"
	// Load credentials from file
	err := LoadCredentials("credentials.txt")
	if err != nil {
		panic("Failed to load credentials: " + err.Error())
	}

	gin.SetMode(gin.ReleaseMode)
	clientWhatsapp := LoginWhatsapp()

	r := gin.Default()
	v1 := r.Group("/api/v1")
	v1.POST("/sendMessage", func(c *gin.Context) {
		SendMessage(c, clientWhatsapp)
	})
	v1.GET("/recvMessage", RecvMessage)
	log.Println("Running on port " + port)
	r.Run(":" + port)
}
