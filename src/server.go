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
  "middleware"
  "controllers"
  "github.com/codegangsta/martini-contrib/binding"
  "models"
)

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

  router := martini.NewRouter()
  // Static Content
  static := martini.Static("public", martini.StaticOptions{
    Fallback:    "/index.html",
    Exclude:     "/api",
    SkipLogging: true,
  })
  router.NotFound(static, http.NotFound)
  m.Use(static)

  jwtMiddleware := middleware.New(middleware.Options{
    ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
      return []byte(conf.JWTSecret), nil
    },
    SigningMethod: jwt.SigningMethodHS256,
    UserProperty:  "user",
  })

  // API Routes
  m.Use(render.Renderer())
  router.Group("/api", func(r martini.Router) {
    r.Get("/friends", jwtMiddleware.CheckJWT, controllers.GetFriends)
    r.Post("/login/:provider", binding.Bind(models.ExtAuthCredentials{}), controllers.LoginWithProvider)
  })
  m.MapTo(router, (*martini.Routes)(nil))
  m.Action(router.Handle)
  m.RunOnAddr("localhost:" + strconv.Itoa(int(conf.ServerPort)))
}
