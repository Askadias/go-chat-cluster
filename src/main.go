package main

import (
  "github.com/go-martini/martini"
  "github.com/Askadias/go-chat-cluster/src/conf"
  "net/http"
  "github.com/codegangsta/martini-contrib/render"
  "github.com/martini-contrib/gzip"
  "github.com/martini-contrib/sessions"
  "strconv"
  "github.com/dgrijalva/jwt-go"
  "github.com/Askadias/go-chat-cluster/src/controllers"
  "github.com/codegangsta/martini-contrib/binding"
  "github.com/Askadias/go-chat-cluster/src/models"
  "github.com/Askadias/go-chat-cluster/src/services"
  "github.com/Askadias/go-chat-cluster/src/middleware"
  "github.com/Askadias/go-chat-cluster/src/middleware/auth"
  "github.com/Askadias/go-chat-cluster/src/db"
)

var jwtOptions = auth.Options{
  ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
    return []byte(conf.JWTSecret), nil
  },
  SigningMethod: jwt.SigningMethodHS256,
  Extractor:     auth.FromJWTCookie,
  UserProperty:  conf.JWTUserPropName,
}

var fbOptions = services.FBOptions{
  ClientId:     conf.FBClientID,
  ClientSecret: conf.FBClientSecret,
  RedirectURL:  conf.FBRedirectURL,
  BaseURL:      conf.FBBaseURL,
  TimeoutMS:    conf.FBTimeoutMS,
}

var chatDBOptions = db.ChatDBOptions{
  MongoURL:       conf.MongoURL,
  MongoDBName:    conf.MongoDBName,
  MongoTimeout:   conf.MongoTimeout,
  MaxChatMembers: conf.MaxChatMembers,
  MaxOpenedChats: conf.MaxOpenedChats,
}

var redisOptions = db.RedisOptions{
  RedisURL: conf.RedisURL,
}

func main() {
  m := martini.New()
  // Add Logging
  m.Use(middleware.Logger())
  // Add Compression
  m.Use(gzip.All())
  m.Use(martini.Recovery())

  // Add Sessions Support
  store := sessions.NewCookieStore([]byte(conf.SessionSecret))
  store.Options(sessions.Options{
    Secure:   true,
    HttpOnly: true,
  })
  m.Use(sessions.Sessions("s", store))

  // Static Content
  static := martini.Static("public", martini.StaticOptions{
    Fallback:    "/index.html",
    Exclude:     "/api",
    SkipLogging: true,
  })
  router := martini.NewRouter()
  router.NotFound(static, http.NotFound)
  m.Use(static)

  jwtMiddleware := auth.NewJwtMiddleware(jwtOptions)

  // Injecting Services
  facebook := services.NewFacebook(fbOptions)
  chat := db.NewMongoChat(chatDBOptions)
  m.MapTo(facebook, (*services.OAuth)(nil))
  m.MapTo(facebook, (*services.Account)(nil))
  m.MapTo(facebook, (*services.Friends)(nil))
  m.MapTo(chat, (*db.Chat)(nil))
  m.MapTo(chat, (*db.ChatLog)(nil))

  // WebSocket Manager
  bus := db.NewRedisBus(redisOptions)
  connectionManager := services.NewConnectionManager(bus, chat)
  m.Map(connectionManager)

  // API Routes
  m.Use(render.Renderer())
  router.Group("/api", func(r martini.Router) {
    r.Get("/friends", jwtMiddleware.CheckJWT, controllers.GetFriends)
    r.Get("/users", jwtMiddleware.CheckJWT, controllers.GetUsers)
    r.Get("/users/:id", jwtMiddleware.CheckJWT, controllers.GetUser)
    r.Get("/rooms", jwtMiddleware.CheckJWT, controllers.GetRooms)
    r.Post("/rooms", binding.Bind(models.Room{}), jwtMiddleware.CheckJWT, controllers.CreateRoom)
    r.Get("/rooms/:id", jwtMiddleware.CheckJWT, controllers.GetRoom)
    r.Delete("/rooms/:id", jwtMiddleware.CheckJWT, controllers.DeleteRoom)
    r.Post("/rooms/:id/members/:memberID", jwtMiddleware.CheckJWT, controllers.AddRoomMember)
    r.Delete("/rooms/:id/members/:memberID", jwtMiddleware.CheckJWT, controllers.RemoveRoomMember)
    r.Get("/rooms/:id/log", jwtMiddleware.CheckJWT, controllers.GetChatLog)
    r.Post("/rooms/:id/log", binding.Bind(models.Message{}), jwtMiddleware.CheckJWT, controllers.SendMessage)
    r.Post("/login/:provider", binding.Bind(models.ExtAuthCredentials{}), controllers.LoginWithProvider)
    r.Get("/ws", jwtMiddleware.CheckJWT, controllers.ConnectToChat)
  })
  m.MapTo(router, (*martini.Routes)(nil))
  m.Action(router.Handle)
  m.RunOnAddr("localhost:" + strconv.Itoa(int(conf.ServerPort)))
}
