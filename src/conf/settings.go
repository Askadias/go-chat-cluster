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
var FBClientID string
var FBClientSecret string
var FBRedirectURL string
var FBTimeoutMS time.Duration
var FBBaseURL string
var FBAuthorizeURL string
var FBScope string

func init() {
  mode := os.Getenv("MARTINI_ENV")
  switch mode {
  case "production":
    FBTimeoutMS = 800
    SessionSecret = os.Getenv("SESSION_SECRET")
    JWTSecret = os.Getenv("JWT_SECRET")
    ServerPort, _ = strconv.ParseInt(os.Getenv("PORT"), 10, 0)
  default:
    FBClientID = "180253089366075"
    FBClientSecret = "e89384a22b77d638b8b2ba4ec1d458e1"
    FBRedirectURL = "http://localhost:3000/authorized"
    FBTimeoutMS = 500
    FBScope = "public_profile,user_friends"
    FBBaseURL = "https://graph.facebook.com/v2.12"
    FBAuthorizeURL = "https://www.facebook.com/v2.12/dialog/oauth"
    SessionSecret = "E4Nkf1ZZ5vRwB5CgvYMDzb12pQ7CU1Tg"
    JWTSecret = "BfqQHegw8cvC22unqNTiIuQVm89jSPLj"
    ServerPort = 3000
  }
}
