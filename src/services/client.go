package services

import (
  "github.com/gorilla/websocket"
  "encoding/json"
  "models"
)

type Client struct {
  Id     string
  Socket *websocket.Conn
  Send   chan []byte
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
    jsonMessage, _ := json.Marshal(&models.Message{From: c.Id, Body: string(message)})
    ChatManager.Broadcast <- jsonMessage
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
