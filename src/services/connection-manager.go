package services

import (
  "models"
  "encoding/json"
  "log"
  "sync"
  "db"
  "conf"
)

type ConnectionManager struct {
  Connections map[string]*Connection
  Register    chan *Connection
  Unregister  chan *Connection
  // --------------------------------------------
  bus   db.Bus
  chat  db.Chat
  mutex sync.RWMutex
}

func NewConnectionManager(bus db.Bus, chat db.Chat) *ConnectionManager {
  manager := &ConnectionManager{
    Register:    make(chan *Connection),
    Unregister:  make(chan *Connection),
    Connections: make(map[string]*Connection),
    bus:         bus,
    chat:        chat,
  }
  go manager.run()
  return manager
}

func (manager *ConnectionManager) run() {
  for {
    select {
    case conn := <-manager.Register:
      if err := manager.bus.Subscribe(conn.UserID); err != nil {
        log.Fatalln("Unable to connect user:", conn.UserID)
      } else {
        manager.mutex.Lock()
        manager.Connections[conn.UserID] = conn
        manager.mutex.Unlock()
        jsonMessage, _ := json.Marshal(&models.Message{Type: "open"})
        manager.send(jsonMessage, conn)
      }
    case conn := <-manager.Unregister:
      if _, ok := manager.Connections[conn.UserID]; ok {
        close(conn.Send)
        manager.mutex.Lock()
        delete(manager.Connections, conn.UserID)
        manager.mutex.Unlock()
        jsonMessage, _ := json.Marshal(&models.Message{Type: "close"})
        manager.send(jsonMessage, conn)
      }
      if err := manager.bus.Unsubscribe(conn.UserID); err != nil {
        log.Println("Unable to disconnect user:", conn.UserID)
      }
    case message := <-manager.bus.Receive():
      for userID, msg := range message {
        manager.Connections[userID].Send <- msg
      }
    }
  }
}

func (manager *ConnectionManager) Broadcast(message *models.Message, auditory []string) *conf.ApiError {
  jsonMessage, _ := json.Marshal(message)
  for _, memberId := range auditory {
    if err := manager.bus.Publish(memberId, jsonMessage); err != nil {
      log.Println("Failed to publish message of type", message.Type, "to user", memberId)
      return conf.ErrBroadcastFailure
    }
  }
  return nil
}

func (manager *ConnectionManager) send(message []byte, ignore *Connection) {
  for _, conn := range manager.Connections {
    if conn != ignore {
      conn.Send <- message
    }
  }
}
