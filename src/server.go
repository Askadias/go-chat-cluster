package main

import (
  "github.com/go-martini/martini"
  "conf"
  "net/http"
  "github.com/codegangsta/martini-contrib/render"
  "github.com/martini-contrib/gzip"
  "github.com/martini-contrib/sessions"
  "strconv"
  "github.com/dgrijalva/jwt-go"
  "controllers"
  "github.com/codegangsta/martini-contrib/binding"
  "models"
  "services"
  "middleware"
  "middleware/auth"
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

var chatDBOptions = services.ChatDBOptions{
  MongoURL:     conf.MongoURL,
  MongoDBName:  conf.MongoDBName,
  MongoTimeout: conf.MongoTimeout,
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

  // WebSocket Manager
  go services.ChatManager.Start()

  // Injecting Services
  facebook := services.NewFacebook(fbOptions)
  chatDBOptions.Facebook = facebook
  chat := services.NewChat(chatDBOptions)
  m.Map(facebook)
  m.Map(chat)

  // API Routes
  m.Use(render.Renderer())
  router.Group("/api", func(r martini.Router) {
    r.Get("/friends", jwtMiddleware.CheckJWT, controllers.GetFriends)
    r.Get("/rooms", jwtMiddleware.CheckJWT, controllers.GetRooms)
    r.Post("/rooms", binding.Bind(models.ChatRoom{}), jwtMiddleware.CheckJWT, controllers.CreateRoom)
    r.Get("/rooms/:id", jwtMiddleware.CheckJWT, controllers.GetRoom)
    r.Delete("/rooms/:id", jwtMiddleware.CheckJWT, controllers.DeleteRoom)
    r.Post("/rooms/:id/members/:memberId", jwtMiddleware.CheckJWT, controllers.AddRoomMember)
    r.Delete("/rooms/:id/members/:memberId", jwtMiddleware.CheckJWT, controllers.RemoveRoomMember)
    r.Get("/rooms/:id/log", jwtMiddleware.CheckJWT, controllers.GetChatLog)
    r.Post("/rooms/:id/log", binding.Bind(models.Message{}), jwtMiddleware.CheckJWT, controllers.LogMessage)
    r.Post("/login/:provider", binding.Bind(models.ExtAuthCredentials{}), controllers.LoginWithProvider)
    r.Get("/ws", jwtMiddleware.CheckJWT, controllers.ConnectToChat)
  })
  m.MapTo(router, (*martini.Routes)(nil))
  m.Action(router.Handle)
  m.RunOnAddr("localhost:" + strconv.Itoa(int(conf.ServerPort)))
}
