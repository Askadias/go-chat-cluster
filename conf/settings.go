package conf

import (
  "os"
  "time"
  "strconv"
)

const (
  SITE_NAME = "Go Chat Cluster"
)

var ServerPort int64
var SessionSecret string
var JWTSecret string
var JWTUserPropName string
var RedisURL string
var MongoURL string
var MongoDBName string
var MongoTimeout time.Duration
var MaxChatMembers int
var MaxOpenedChats int
var FBClientID string
var FBClientSecret string
var FBRedirectURL string
var FBTimeoutMS time.Duration
var FBBaseURL string

var ChatLogLimit int

func init() {
  mode := os.Getenv("MARTINI_ENV")
  switch mode {
  case "production":
    FBTimeoutMS = 2000
    SessionSecret = os.Getenv("SESSION_SECRET")
    JWTSecret = os.Getenv("JWT_SECRET")
    ServerPort, _ = strconv.ParseInt(os.Getenv("PORT"), 10, 0)
  default:
    RedisURL = os.Getenv("REDIS_URL")
    MongoURL = os.Getenv("MONGO_URL")
    MongoDBName = os.Getenv("MONGO_DB_NAME")
    MongoTimeout = 1 * time.Second
    MaxChatMembers = 100
    MaxOpenedChats = 10
    FBClientID = "180253089366075"
    FBClientSecret = "e89384a22b77d638b8b2ba4ec1d458e1"
    FBRedirectURL = "http://localhost:3000/authorized"
    FBTimeoutMS = 500
    FBBaseURL = "https://graph.facebook.com/v2.12"
    SessionSecret = "E4Nkf1ZZ5vRwB5CgvYMDzb12pQ7CU1Tg"
    JWTSecret = "BfqQHegw8cvC22unqNTiIuQVm89jSPLj"
    JWTUserPropName = "user"
    ServerPort = 3000
    ChatLogLimit = 10
  }
}
