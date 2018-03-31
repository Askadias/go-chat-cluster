package services

import (
  "github.com/gorilla/websocket"
  "encoding/json"
  "models"
  "log"
  "db"
)

type Client struct {
  Id      string
  Socket  *websocket.Conn
  Send    chan []byte
  ChatLog db.ChatLog
}

func (c *Client) Read() {
  defer func() {
    ChatManager.Unregister <- c
    c.Socket.Close()
  }()

  for {
    _, message, err := c.Socket.ReadMessage()
    if err != nil {
      ChatManager.Unregister <- c
      c.Socket.Close()
      break
    }
    log.Println("Message received:", string(message))
    msg := models.Message{}
    json.Unmarshal(message, &msg)
    msg.From = c.Id
    jsonMessage, _ := json.Marshal(&msg)
    ChatManager.Broadcast <- jsonMessage
    if msg.Type != "update" {
      c.ChatLog.AddMessage(msg)
    }
  }
}

func (c *Client) Write() {
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
