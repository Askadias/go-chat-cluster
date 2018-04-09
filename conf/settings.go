package conf

import (
  "time"
  "github.com/kelseyhightower/envconfig"
)

var System SystemConf

type SystemConf struct {
  // Application host
  Host string
  // Application port (3000 by default)
  Port int `default:"3000"`
  // HTTP Session encryption secret key
  SessionSecret string `default:"E4Nkf1ZZ5vRwB5CgvYMDzb12pQ7CU1Tg"`
  // User Authentication JWT encryption key
  JWTSecret string `default:"BfqQHegw8cvC22unqNTiIuQVm89jSPLj"`
  // User property name to store in the request context
  JWTUserPropName string `default:"user"`
}

var Socket SocketConf

type SocketConf struct {
  // Time allowed to write a message to the peer.
  WriteWait time.Duration `default:"10s",split_words:"true"`
  // Time allowed to read the next pong message from the peer.
  PongWait time.Duration `default:"60s",split_words:"true"`
  // Send pings to peer with this period. Must be less than pongWait.
  PingPeriod time.Duration `default:"54s",split_words:"true"`
  // Maximum message size allowed from peer.
  MaxMessageSize int64 `default:"1024",split_words:"true"`
  // HandshakeTimeout specifies the duration for the handshake to complete.
  HandshakeTimeout time.Duration `default:"10s",split_words:"true"`
  // ReadBufferSize and WriteBufferSize specify I/O buffer sizes. If a buffer
  // size is zero, then buffers allocated by the HTTP server are used. The
  // I/O buffer sizes do not limit the size of the messages that can be sent
  // or received.
  ReadBufferSize  int `default:"1024",split_words:"true"`
  WriteBufferSize int `default:"1024",split_words:"true"`
  // Goroutines pool size for receiving and broadcasting messages
  SendersPoolSize   int `default:"128",split_words:"true"`
  ReceiversPoolSize int `default:"128",split_words:"true"`
}

var Facebook FacebookConf

type FacebookConf struct {
  // Facebook application client ID
  ClientID string `default:"180253089366075"`
  // Facebook application client secret
  ClientSecret string
  // Authorization URL registered in facebook as allowed for redirection
  RedirectURL string `default:"http://localhost:3000/authorized"`
  // Facebook request timeout
  Timeout time.Duration `default:"5s"`
  // Base url for the facebook graph API
  BaseURL string `default:"https://graph.facebook.com/v2.12"`
}

var Chat ChatConf

type ChatConf struct {
  // Maximum number of members allowed for a single room
  MaxMembers int `default:"100"`
  // Maximum number of opened chat rooms per user
  MaxOpenedRooms int `default:"10"`
  // Default limit of messages to be returned with a single request
  DefaultMessagesLimit int `default:"10"`
}

var Redis RedisConf

type RedisConf struct {
  URL string
  // Maximum number of idle connections in the pool.
  MaxIdle int `default:"16"`
  // Maximum number of connections allocated by the pool at a given time.
  // When zero, there is no limit on the number of connections in the pool.
  MaxActive int `default:"16"`
  // Close connections after remaining idle for this duration. If the value
  // is zero, then idle connections are not closed. Applications should set
  // the timeout to a value less than the server's timeout.
  IdleTimeout time.Duration `default:"1s"`
  // Close connections older than this duration. If the value is zero, then
  // the pool does not close connections based on age.
  MaxConnLifetime time.Duration
  // Cache expiration duration
  CacheTTL time.Duration `default:"5m"`
}

var Mongo MongoConf

type MongoConf struct {
  // Full MongoDB url
  URL string
  // Mongo Database Name
  DBName string `default:"go-chat-cluster"`
  // Operation timeout
  Timeout time.Duration `default:"1s"`
  // Collection name for storing chat room info
  RoomCollectionName string `default:"rooms"`
  // Collection name for storing chat log
  MessagesCollectionName string `default:"messages"`
}

func init() {
  err := envconfig.Process("", &System)
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
