package services

import (
  "github.com/Askadias/go-chat-cluster/models"
  "encoding/json"
  "log"
  "sync"
  "github.com/Askadias/go-chat-cluster/db"
  "github.com/Askadias/go-chat-cluster/conf"
  "time"
)

type ConnectionManager struct {
  Connections map[string]*Connection
  Register    chan *Connection
  Unregister  chan *Connection
  socketConf  conf.SocketConf
  // --------------------------------------------
  bus   db.Bus
  chat  db.Chat
  mutex sync.RWMutex
}

func NewConnectionManager(bus db.Bus, chat db.Chat, socketConf conf.SocketConf) *ConnectionManager {
  manager := &ConnectionManager{
    Register:    make(chan *Connection),
    Unregister:  make(chan *Connection),
    Connections: make(map[string]*Connection),
    socketConf:  socketConf,
    bus:         bus,
    chat:        chat,
  }
  go manager.broadcasting()
  go manager.ping()
  return manager
}

func (manager *ConnectionManager) broadcasting() {
  for {
    select {
    case conn := <-manager.Register:
      if err := manager.bus.Subscribe(conn.UserID); err != nil {
        log.Println("Unable to connect user:", conn.UserID)
      } else {
        manager.mutex.Lock()
        manager.Connections[conn.UserID] = conn
        manager.mutex.Unlock()
      }
    case conn := <-manager.Unregister:
      if _, ok := manager.Connections[conn.UserID]; ok {
        manager.mutex.Lock()
        delete(manager.Connections, conn.UserID)
        manager.mutex.Unlock()
      }
      if err := manager.bus.Unsubscribe(conn.UserID); err != nil {
        log.Println("Unable to disconnect user:", conn.UserID)
      }
    case message := <-manager.bus.Receive():
      manager.mutex.RLock()
      for userID, msg := range message {
        if err := manager.Connections[userID].Send(msg); err != nil {
          log.Println("Unable to send message to user", userID)
          manager.Unregister <- manager.Connections[userID]
        }
      }
      manager.mutex.RUnlock()
    }
  }
}

func (manager *ConnectionManager) ping() {
  ticker := time.NewTicker(manager.socketConf.PingPeriod)
  defer func() {
    ticker.Stop()
  }()
  for {
    select {
    case <-ticker.C:
      for userID, connection := range manager.Connections {
        if err := connection.Ping(); err != nil {
          log.Println("Failed to ping connection for user", userID)
          manager.Unregister <- connection
        }
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
