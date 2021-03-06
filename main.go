package main

import (
  "github.com/Askadias/go-chat-cluster/conf"
  "github.com/Askadias/go-chat-cluster/controllers"
  "github.com/Askadias/go-chat-cluster/db"
  "github.com/Askadias/go-chat-cluster/middleware"
  "github.com/Askadias/go-chat-cluster/middleware/auth"
  "github.com/Askadias/go-chat-cluster/models"
  "github.com/Askadias/go-chat-cluster/services"
  "github.com/codegangsta/martini-contrib/binding"
  "github.com/codegangsta/martini-contrib/render"
  "github.com/dgrijalva/jwt-go"
  "github.com/go-martini/martini"
  "github.com/martini-contrib/gzip"
  "github.com/martini-contrib/sessions"
  "net/http"
  "strconv"
)

var jwtOptions = auth.Options{
  ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
    return []byte(conf.System.JWTSecret), nil
  },
  SigningMethod: jwt.SigningMethodHS256,
  Extractor:     auth.FromJWTCookie,
  UserProperty:  conf.System.JWTUserPropName,
}

var redisPool = db.NewRedisPool(conf.Redis)
var redisBusOptions = db.RedisBusOptions{
  RedisPool: redisPool,
}
var redisCacheOptions = db.RedisCacheOptions{
  RedisPool: redisPool,
  CacheTTL: conf.Redis.CacheTTL,
}

func main() {
  m := martini.New()
  // Add Logging
  m.Use(middleware.Metadata())
  m.Use(middleware.Logger())
  // Add Compression
  m.Use(gzip.All())
  m.Use(martini.Recovery())

  // Add Sessions Support
  store := sessions.NewCookieStore([]byte(conf.System.SessionSecret))
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

  // WebSocket Manager
  bus := db.NewRedisBus(redisBusOptions)
  connectionManager := services.NewConnectionManager(bus, conf.Socket)
  m.Map(connectionManager)

  // Injecting Services
  cache := db.NewRedisCache(redisCacheOptions)
  facebook := services.NewFacebook(conf.Facebook, &http.Client{
    Timeout: conf.Facebook.Timeout,
  })
  chatDB := db.NewMongoChat(conf.Mongo)
  chat := services.NewChat(conf.Chat, chatDB, chatDB, chatDB, cache, connectionManager)
  m.Map(chat)
  m.MapTo(facebook, (*services.OAuth)(nil))
  m.MapTo(facebook, (*services.Account)(nil))
  m.MapTo(facebook, (*services.Friends)(nil))

  // API Routes
  m.Use(render.Renderer())
  router.Group("/api", func(r martini.Router) {
    r.Get("/friends", jwtMiddleware.CheckJWT, controllers.GetFriends)
    r.Get("/friends/permissions", jwtMiddleware.CheckJWT, controllers.HasFriendsPermissions)
    r.Get("/users", jwtMiddleware.CheckJWT, controllers.GetUsers)
    r.Get("/users/:id", jwtMiddleware.CheckJWT, controllers.GetUser)
    r.Get("/rooms", jwtMiddleware.CheckJWT, controllers.GetRooms)
    r.Post("/rooms", binding.Bind(models.Room{}), jwtMiddleware.CheckJWT, controllers.CreateRoom)
    r.Get("/rooms/:id", jwtMiddleware.CheckJWT, controllers.GetRoom)
    r.Delete("/rooms/:id", jwtMiddleware.CheckJWT, controllers.DeleteRoom)
    r.Get("/rooms/members/info", jwtMiddleware.CheckJWT, controllers.GetAllRoomsInfo)
    r.Get("/rooms/:id/members/info", jwtMiddleware.CheckJWT, controllers.GetAllMembersInfo)
    r.Get("/rooms/:id/members/:memberID/info", jwtMiddleware.CheckJWT, controllers.GetMemberInfo)
    r.Put("/rooms/:id/members/:memberID/info", jwtMiddleware.CheckJWT, controllers.UpdateMemberLastReadTime)
    r.Post("/rooms/:id/members/:memberID", jwtMiddleware.CheckJWT, controllers.AddRoomMember)
    r.Delete("/rooms/:id/members/:memberID", jwtMiddleware.CheckJWT, controllers.RemoveRoomMember)
    r.Get("/rooms/:id/log", jwtMiddleware.CheckJWT, controllers.GetChatLog)
    r.Post("/rooms/:id/log", binding.Bind(models.Message{}), jwtMiddleware.CheckJWT, controllers.SendMessage)
    r.Post("/login/:provider", binding.Bind(models.ExtAuthCredentials{}), controllers.LoginWithProvider)
    r.Get("/ws", jwtMiddleware.CheckJWT, controllers.ConnectToChat)
  })
  m.MapTo(router, (*martini.Routes)(nil))
  m.Action(router.Handle)
  m.RunOnAddr(conf.System.Host + ":" + strconv.Itoa(conf.System.Port))
}
