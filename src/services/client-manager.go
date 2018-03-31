package services

import (
  "models"
  "encoding/json"
)

var ChatManager = NewClientManager()

type ClientManager struct {
  Clients    map[*Client]bool
  Broadcast  chan []byte
  Register   chan *Client
  Unregister chan *Client
}

func NewClientManager() *ClientManager {
  return &ClientManager{
    Broadcast:  make(chan []byte),
    Register:   make(chan *Client),
    Unregister: make(chan *Client),
    Clients:    make(map[*Client]bool),
  }
}

func (manager *ClientManager) Start() {
  for {
    select {
    case conn := <-manager.Register:
      manager.Clients[conn] = true
      jsonMessage, _ := json.Marshal(&models.Message{Type: "open"})
      manager.send(jsonMessage, conn)
    case conn := <-manager.Unregister:
      if _, ok := manager.Clients[conn]; ok {
        close(conn.Send)
        delete(manager.Clients, conn)
        jsonMessage, _ := json.Marshal(&models.Message{Type: "close"})
        manager.send(jsonMessage, conn)
      }
    case message := <-manager.Broadcast:
      for conn := range manager.Clients {
        select {
        case conn.Send <- message:
        default:
          close(conn.Send)
          delete(manager.Clients, conn)
        }
      }

    }
  }
}

func (manager *ClientManager) send(message []byte, ignore *Client) {
  for conn := range manager.Clients {
    if conn != ignore {
      conn.Send <- message
    }
  }
}
