package middleware

import (
  "github.com/go-martini/martini"
  "net/http"
  "github.com/satori/go.uuid"
  "time"
)

const RequestIdHeader = "X-Request-ID"
const DeviceIdCookie = "DeviceID"

type DeviceID string

type RequestID string

func Metadata() martini.Handler {
  return func(c martini.Context, req *http.Request, w http.ResponseWriter) {
    requestID := req.Header.Get(RequestIdHeader)
    deviceID := ""
    if deviceIDCookie, err := req.Cookie(DeviceIdCookie); err == nil {
      deviceID = deviceIDCookie.Value
    }

    if requestID == "" {
      requestID = generateID()
      req.Header.Set(RequestIdHeader, requestID)
    }
    if deviceID == "" {
      deviceID = generateID()
      deviceIdCookie := http.Cookie{Name: DeviceIdCookie, Value:deviceID, Expires: time.Now().Add(365 * 24 * time.Hour)}
      http.SetCookie(w, &deviceIdCookie)
    }
    c.Map(RequestID(requestID))
    c.Map(DeviceID(deviceID))
  }
}

func generateID() string {
  if id, err := uuid.NewV4(); err != nil {
    return ""
  } else {
    return id.String()
  }
}
