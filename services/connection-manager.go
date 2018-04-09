package services

import (
  "github.com/Askadias/go-chat-cluster/models"
  "encoding/json"
  "log"
  "sync"
  "github.com/Askadias/go-chat-cluster/db"
  "github.com/Askadias/go-chat-cluster/conf"
  "time"
  "github.com/Jeffail/tunny"
)

type BroadcastPackage struct {
  Message  *models.Message
  Auditory []string
}

type MessageJob struct {
  memberID string
  message  []byte
}

type ConnectionManager struct {
  Connections map[string]*Connection
  Register    chan *Connection
  Unregister  chan *Connection
  Broadcast   chan *BroadcastPackage
  socketConf  conf.SocketConf
  // --------------------------------------------
  bus   db.Bus
  chat  db.Chat
  mutex sync.RWMutex
}

func NewConnectionManager(bus db.Bus, chat db.Chat, socketConf conf.SocketConf) *ConnectionManager {
  m := &ConnectionManager{
    Register:    make(chan *Connection),
    Unregister:  make(chan *Connection),
    Broadcast:   make(chan *BroadcastPackage),
    Connections: make(map[string]*Connection),
    socketConf:  socketConf,
    bus:         bus,
    chat:        chat,
  }
  go m.broadcasting()
  go m.ping()
  return m
}

func (m *ConnectionManager) broadcasting() {
  sendersPool := tunny.NewFunc(m.socketConf.SendersPoolSize, func(messageJob interface{}) interface{} {
    job := messageJob.(MessageJob)
    if err := m.bus.Publish(job.memberID, job.message); err != nil {
      log.Println("Failed to publish message for user", job.memberID)
    }
    return nil
  })
  receiversPool := tunny.NewFunc(m.socketConf.ReceiversPoolSize, func(messageJob interface{}) interface{} {
    job := messageJob.(MessageJob)
    m.mutex.RLock()
    if conn, ok := m.Connections[job.memberID]; ok {
      if err := conn.Send(job.message); err != nil {
        log.Println("Unable to deliver message to user", job.memberID)
        m.Unregister <- m.Connections[job.memberID]
      }
    } else {
      log.Println("Unable to deliver message to disconnected user", job.memberID)
    }
    m.mutex.RUnlock()
    return nil
  })
  defer sendersPool.Close()
  defer receiversPool.Close()
  for {
    select {
    case conn := <-m.Register:
      if err := m.bus.Subscribe(conn.UserID); err != nil {
        log.Println("Unable to connect user:", conn.UserID)
      } else {
        m.mutex.Lock()
        m.Connections[conn.UserID] = conn
        m.mutex.Unlock()
      }
    case conn := <-m.Unregister:
      if _, ok := m.Connections[conn.UserID]; ok {
        m.mutex.Lock()
        delete(m.Connections, conn.UserID)
        m.mutex.Unlock()
      }
      if err := m.bus.Unsubscribe(conn.UserID); err != nil {
        log.Println("Unable to disconnect user:", conn.UserID)
      }
    case pack := <-m.Broadcast:
      jsonMessage, _ := json.Marshal(pack.Message)
      for _, memberID := range pack.Auditory {
        sendersPool.Process(MessageJob{memberID, jsonMessage})
      }
    case message := <-m.bus.Receive():
      for memberID, msg := range message {
        receiversPool.Process(MessageJob{memberID, msg})
      }
    }
  }
}

func (m *ConnectionManager) ping() {
  ticker := time.NewTicker(m.socketConf.PingPeriod)
  defer func() {
    ticker.Stop()
  }()
  for {
    select {
    case <-ticker.C:
      for userID, connection := range m.Connections {
        if err := connection.Ping(); err != nil {
          log.Println("Failed to ping connection for user", userID)
          m.Unregister <- connection
        }
      }
    }
  }
}
