package services

import (
  "github.com/gorilla/websocket"
  "encoding/json"
  "models"
  "log"
  "db"
)

type Connection struct {
  UserID  string
  Socket  *websocket.Conn
  Send    chan []byte
  ChatLog db.ChatLog
  Manager ConnectionManager
}

func (c *Connection) Read() {
  defer func() {
    c.Manager.Unregister <- c
    c.Socket.Close()
  }()

  for {
    _, message, err := c.Socket.ReadMessage()
    if err != nil {
      c.Manager.Unregister <- c
      c.Socket.Close()
      break
    }
    log.Println("Message received:", string(message))
    msg := models.Message{}
    json.Unmarshal(message, &msg)
    msg.From = c.UserID

    c.Manager.Broadcast <- msg
    if msg.Type != "update" {
      c.ChatLog.AddMessage(msg)
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
