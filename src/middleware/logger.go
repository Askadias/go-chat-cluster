package middleware

import (
  "log"
  "net/http"
  "time"
  "github.com/go-martini/martini"
  "github.com/satori/go.uuid"
  "reflect"
)

// Logger returns a middleware handler that logs the request as it goes in and the response as it goes out.
func Logger() martini.Handler {
  return func(res http.ResponseWriter, req *http.Request, c martini.Context, log *log.Logger) {
    start := time.Now()

    addr := req.Header.Get("X-Real-IP")
    if addr == "" {
      addr = req.Header.Get("X-Forwarded-For")
      if addr == "" {
        addr = req.RemoteAddr
      }
    }
    requestId, _ := uuid.NewV4()
    // TODO check that rqID is stored in the context or inject a custom context
    c.Set(reflect.TypeOf("rqID"), reflect.ValueOf(requestId.String()))

    log.Printf("%s - %s %s \"%s %s\"",
      addr,
      time.Now().UTC().String(),
      requestId.String(),
      req.Method,
      req.URL.Path)

    rw := res.(martini.ResponseWriter)

    c.Next()
    log.Printf("%s - %s %s %d \"%s %s\" %s",
      addr,
      time.Now().UTC().String(),
      requestId.String(),
      rw.Status(),
      req.Method,
      req.URL.Path,
      time.Since(start))
  }
}
