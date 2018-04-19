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
  "github.com/gorilla/websocket"
  "github.com/Askadias/go-chat-cluster/middleware"
)

// Connection DTO for channel
type Connection struct {
  UserID   string              // abstract user ID
  DeviceID middleware.DeviceID // user device ID to support chats opened on several devices
  Socket   *websocket.Conn     // socket connection for incoming messages and notifications
}

type BroadcastPackage struct {
  Message  *models.Message
  Auditory []string
}

type MessageJob struct {
  memberID string
  message  []byte
}

type ConnectionManager struct {
  Connections map[string]map[middleware.DeviceID]*websocket.Conn
  Register    chan *Connection
  Unregister  chan *Connection
  Broadcast   chan *BroadcastPackage
  socketConf  conf.SocketConf
  // --------------------------------------------
  bus   db.Bus
  mutex sync.RWMutex
}

func NewConnectionManager(bus db.Bus, socketConf conf.SocketConf) *ConnectionManager {
  m := &ConnectionManager{
    Register:    make(chan *Connection),
    Unregister:  make(chan *Connection),
    Broadcast:   make(chan *BroadcastPackage),
    Connections: make(map[string]map[middleware.DeviceID]*websocket.Conn),
    socketConf:  socketConf,
    bus:         bus,
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
    if devices, ok := m.Connections[job.memberID]; ok {
      for deviceID, socket := range devices {
        if err := m.sendMessage(socket, job.message); err != nil {
          log.Println("Unable to deliver message to user", job.memberID)
          m.Unregister <- &Connection{job.memberID, deviceID, socket}
        }
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
        if _, ok := m.Connections[conn.UserID]; !ok {
          m.Connections[conn.UserID] = make(map[middleware.DeviceID]*websocket.Conn)
        }
        m.Connections[conn.UserID][conn.DeviceID] = conn.Socket
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
      for userID, devices := range m.Connections {
        for deviceID, socket := range devices {
          if err := m.pingSocket(socket); err != nil {
            log.Println("Failed to ping connection for user", userID)
            m.Unregister <- &Connection{userID, deviceID, socket}
          }
        }
      }
    }
  }
}

func (m *ConnectionManager) sendMessage(socket *websocket.Conn, message []byte) error {
  socket.SetWriteDeadline(time.Now().Add(conf.Socket.WriteWait))
  return socket.WriteMessage(websocket.TextMessage, message)
}

func (m *ConnectionManager) pingSocket(socket *websocket.Conn) error {
  socket.SetWriteDeadline(time.Now().Add(conf.Socket.WriteWait))
  return socket.WriteMessage(websocket.PingMessage, []byte{})
}
