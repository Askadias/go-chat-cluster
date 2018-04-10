package services

import (
  "github.com/gorilla/websocket"
  "time"
  "github.com/Askadias/go-chat-cluster/conf"
)

type Connection struct {
  UserID string
  Socket *websocket.Conn
}

func NewConnection(userID string, socket *websocket.Conn) *Connection {
  connection := &Connection{
    UserID: userID,
    Socket: socket,
  }
  return connection
}

func (c *Connection) Send(message []byte) error {
  c.Socket.SetWriteDeadline(time.Now().Add(conf.Socket.WriteWait))
  return c.Socket.WriteMessage(websocket.TextMessage, message)
}

func (c *Connection) Ping() error {
  c.Socket.SetWriteDeadline(time.Now().Add(conf.Socket.WriteWait))
  return c.Socket.WriteMessage(websocket.PingMessage, []byte{})
}
