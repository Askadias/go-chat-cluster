package services

import (
  "github.com/gorilla/websocket"
  "github.com/Askadias/go-chat-cluster/db"
)

type Connection struct {
  UserID  string
  Socket  *websocket.Conn
  Send    chan []byte
  ChatLog db.ChatLog
  Manager ConnectionManager
}

func NewConnection(userID string, socket *websocket.Conn, chatLog db.ChatLog, manager ConnectionManager) *Connection {
  connection := &Connection{
    UserID:  userID,
    Socket:  socket,
    Send:    make(chan []byte),
    ChatLog: chatLog,
    Manager: manager,
  }
  manager.Register <- connection
  return connection
}

func (c *Connection) Read() {
  defer func() {
    c.Manager.Unregister <- c
    c.Socket.Close()
  }()

  for {
    _, _, err := c.Socket.ReadMessage()
    if err != nil {
      c.Manager.Unregister <- c
      c.Socket.Close()
      break
    }
  }
}

func (c *Connection) Write() {
  defer func() {
    c.Socket.Close()
  }()

  for {
    select {
    case message, ok := <-c.Send:
      if !ok {
        c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
        return
      }
      c.Socket.WriteMessage(websocket.TextMessage, message)
    }
  }
}
