// filePath:  store/store.go
package store

import (
	"encoding/json"
	"fmt"
	"kai/kai"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Store is a struct for storing global data and managing WebSocket connections
type Store struct {
	OrdIdx  *kai.BTOrdIdx            // Index for orders
	RecIdx  *kai.BTRecIdx            // Index for records
	clients map[*websocket.Conn]bool // Active WebSocket connections

	// Upgrader for WebSocket connections, allows for custom configurations
	upgrader websocket.Upgrader
}

// instance is the private instance of Store
var instance *Store

// once ensures the singleton instance is created only once
var once sync.Once

// Instance returns the single instance of Store
func Instance() *Store {
	once.Do(func() {
		instance = &Store{
			clients: make(map[*websocket.Conn]bool),
			upgrader: websocket.Upgrader{
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
				CheckOrigin: func(r *http.Request) bool {
					return true // Allow all CORS requests, be stricter in production environments
				},
			},
		}
	})
	return instance
}

// Init initializes the data of the Store and starts the WebSocket server on the specified port
func (s *Store) Init(OrdIdx *kai.BTOrdIdx, RecIdx *kai.BTRecIdx, port string) {
	s.OrdIdx = OrdIdx
	s.RecIdx = RecIdx
	go s.startWebSocketServer(port) // Start the WebSocket server in a new goroutine
	fmt.Println("Listening Socket Port =", port)
}

// AllSocketPush pushes a string message to all connected WebSocket clients
func (s *Store) AllSocketPush(message string) {
	msg := []byte(message)
	for client := range s.clients {
		if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("WebSocket error: %v", err)
			client.Close()
			delete(s.clients, client)
		}
	}
}

func (s *Store) WriteRecorderPush(address, rectype, msg string) (err error) {

	err = s.RecIdx.WriteRecorder(address, rectype, msg)
	if err != nil {
		return err
	}

	// // Construct the JSON message
	// jsonMessage := fmt.Sprintf(`{"address":"%s","rectype":"%s","message":"%s"}`, address, rectype, msg)
	// //store.AllSocketPush(string(message)) // Push the message to all connected WebSocket clients
	// Construct the JSON message
	jsonMessage, err := json.Marshal(gin.H{
		"address": address,
		"rectype": rectype,
		"message": string(msg),
	})
	if err != nil {
		return err
	}
	s.AllSocketPush(string(jsonMessage)) // Push the JSON message to all connected WebSocket clients

	//s.AllSocketPush(jsonMessage) // Push the JSON message to all connected WebSocket clients
	return nil
}

func (s *Store) GoWriteRecorderPush() {
	for {
		address, rectype, msg := s.OrdIdx.SelectPush()
		fmt.Printf("GoWriteRecorderPush() - Address: %s, RecType: %s, Msg: %s\n", address, rectype, msg)

		err := s.WriteRecorderPush(address, rectype, string(msg))
		if err != nil {
			return
		}

	}
}

// startWebSocketServer starts the WebSocket server on a specified port
func (s *Store) startWebSocketServer(port string) {
	http.HandleFunc("/ws", s.handleConnections)
	log.Printf("WebSocket server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}

// handleConnections upgrades HTTP to WebSocket and manages connection lifecycle
func (s *Store) handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()
	s.clients[ws] = true // Register new client

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			delete(s.clients, ws)
			break
		}
	}
}
