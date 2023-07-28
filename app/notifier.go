package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

type Message struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type Notifier struct {
	connections map[*websocket.Conn]struct{}
}

func NewNotifier() *Notifier {

	return &Notifier{
		connections: make(map[*websocket.Conn]struct{}),
	}
}

func (notifier *Notifier) AddConnection(conn *websocket.Conn) {
	notifier.connections[conn] = struct{}{}
	fmt.Printf("Add connection, %d connection on this Process\n", len(notifier.connections))
}

func (notifier *Notifier) RemoveConnection(conn *websocket.Conn) error {
	fmt.Println("Remove Connection")

	if _, ok := notifier.connections[conn]; ok {
		delete(notifier.connections, conn)
		conn.Close()
		return nil
	}

	return errors.New("connection not found")
}
func (notifier *Notifier) Push(messageJSON []byte) {

	var wg sync.WaitGroup

	for conn := range notifier.connections {
		wg.Add(1)
		go func(conn *websocket.Conn) {
			defer wg.Done()

			conn.WriteMessage(websocket.TextMessage, messageJSON)

		}(conn)
	}

	wg.Wait()

}

func (notifier *Notifier) CreateMessage(content string) ([]byte, error) {
	message := Message{
		Content:   content,
		Timestamp: time.Now(),
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message to JSON: %v", err)
	}

	return messageJSON, nil
}
