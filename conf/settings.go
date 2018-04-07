package conf

import (
  "time"
  "github.com/kelseyhightower/envconfig"
  "strconv"
  "os"
)

var System SystemConf

type SystemConf struct {
  // Application host
  Host string
  // Application port (3000 by default)
  Port int
  // HTTP Session encryption secret key
  SessionSecret string
  // User Authentication JWT encryption key
  JWTSecret string
  // User property name to store in the request context
  JWTUserPropName string
}

var Socket SocketConf

type SocketConf struct {
  // Time allowed to write a message to the peer.
  WriteWait time.Duration
  // Time allowed to read the next pong message from the peer.
  PongWait time.Duration
  // Send pings to peer with this period. Must be less than pongWait.
  PingPeriod time.Duration
  // Maximum message size allowed from peer.
  MaxMessageSize int64
  // HandshakeTimeout specifies the duration for the handshake to complete.
  HandshakeTimeout time.Duration
  // ReadBufferSize and WriteBufferSize specify I/O buffer sizes. If a buffer
  // size is zero, then buffers allocated by the HTTP server are used. The
  // I/O buffer sizes do not limit the size of the messages that can be sent
  // or received.
  ReadBufferSize, WriteBufferSize int
}

var Facebook FacebookConf

type FacebookConf struct {
  // Facebook application client ID
  ClientID string
  // Facebook application client secret
  ClientSecret string
  // Authorization URL registered in facebook as allowed for redirection
  RedirectURL string
  // Facebook request timeout
  TimeoutMS time.Duration
  // Base url for the facebook graph API
  BaseURL string
}

var Chat ChatConf

type ChatConf struct {
  // Maximum number of members allowed for a single room
  MaxMembers int
  // Maximum number of opened chat rooms per user
  MaxOpenedRooms int
  // Default limit of messages to be returned with a single request
  DefaultMessagesLimit int
}

var Redis RedisConf

type RedisConf struct {
  URL string
  // Maximum number of idle connections in the pool.
  MaxIdle int
  // Maximum number of connections allocated by the pool at a given time.
  // When zero, there is no limit on the number of connections in the pool.
  MaxActive int
  // Close connections after remaining idle for this duration. If the value
  // is zero, then idle connections are not closed. Applications should set
  // the timeout to a value less than the server's timeout.
  IdleTimeout time.Duration
  // Close connections older than this duration. If the value is zero, then
  // the pool does not close connections based on age.
  MaxConnLifetime time.Duration
}

var Mongo MongoConf

type MongoConf struct {
  // Full MongoDB url
  URL string
  // Mongo Database Name
  DBName string
  // Operation timeout
  Timeout time.Duration
  // Collection name for storing chat room info
  RoomCollectionName string
  // Collection name for storing chat log
  MessagesCollectionName string
}

func init() {
  if port, err := strconv.ParseInt(os.Getenv("PORT"), 10, 0); err != nil || port == 0 {
    System.Port = 3000
  } else {
    System.Port = int(port)
  }
  System.SessionSecret = "E4Nkf1ZZ5vRwB5CgvYMDzb12pQ7CU1Tg"
  System.JWTSecret = "BfqQHegw8cvC22unqNTiIuQVm89jSPLj"
  System.JWTUserPropName = "user"

  Socket.WriteWait = 10 * time.Second
  Socket.PongWait = 60 * time.Second
  Socket.PingPeriod = (Socket.PongWait * 9) / 10
  Socket.HandshakeTimeout = 10 * time.Second
  Socket.MaxMessageSize = 1024
  Socket.ReadBufferSize = 1024
  Socket.WriteBufferSize = 1024

  Facebook.ClientID = "180253089366075"
  Facebook.BaseURL = "https://graph.facebook.com/v2.12"
  Facebook.RedirectURL = "http://localhost:3000/authorized"
  Facebook.TimeoutMS = 10000

  Redis.IdleTimeout = 1 * time.Second
  Redis.MaxActive = 16
  Redis.MaxIdle = 16

  Mongo.Timeout = 1 * time.Second
  Mongo.DBName = "go-chat-cluster"
  Mongo.RoomCollectionName = "rooms"
  Mongo.MessagesCollectionName = "messages"

  Chat.MaxMembers = 100
  Chat.MaxOpenedRooms = 10
  Chat.DefaultMessagesLimit = 10

  err := envconfig.Process("System", &System)
  if err != nil {
    panic("Unable to load 'System' config: " + err.Error())
  }

  err = envconfig.Process("Socket", &Socket)
  if err != nil {
    panic("Unable to load 'Socket' config: " + err.Error())
  }

  err = envconfig.Process("Facebook", &Facebook)
  if err != nil {
    panic("Unable to load 'Facebook' config: " + err.Error())
  }

  err = envconfig.Process("Chat", &Chat)
  if err != nil {
    panic("Unable to load 'Chat' config: " + err.Error())
  }

  err = envconfig.Process("Redis", &Redis)
  if err != nil {
    panic("Unable to load 'Redis' config: " + err.Error())
  }

  err = envconfig.Process("Mongo", &Mongo)
  if err != nil {
    panic("Unable to load 'Mongo' config: " + err.Error())
  }
}
