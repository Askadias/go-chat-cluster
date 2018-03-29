package controllers

import (
  "net/http"
  "github.com/gorilla/websocket"
  "services"
  "github.com/dgrijalva/jwt-go"
  "github.com/codegangsta/martini-contrib/render"
  "conf"
  "log"
  "models"
  "gopkg.in/mgo.v2/bson"
  "time"
  "github.com/go-martini/martini"
  "strconv"
)

// Initializes a WebSocket connection for the current user
func ConnectToChat(res http.ResponseWriter, req *http.Request, render render.Render) {
  tkn := req.Context().Value(conf.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"]

    if conn, err := (&websocket.Upgrader{}).Upgrade(res, req, nil); err != nil {
      http.NotFound(res, req)
    } else {
      client := &services.Client{Id: profileID.(string), Socket: conn, Send: make(chan []byte)}
      services.ChatManager.Register <- client

      go client.Read()
      go client.Write()
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Create chat room
func CreateRoom(room models.ChatRoom, req *http.Request, render render.Render, chat *services.Chat) {
  tkn := req.Context().Value(conf.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)

    log.Println("Creating chat room for", profileID)

    if newRoom, err := chat.CreateRoom(profileID, room); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      render.JSON(http.StatusOK, newRoom)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Returns all the current user's chat rooms
func GetRooms(req *http.Request, render render.Render, chat *services.Chat) {
  tkn := req.Context().Value(conf.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    if rooms, err := chat.GetRooms(profileID); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      render.JSON(http.StatusOK, rooms)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Returns chat room by its ID
func GetRoom(params martini.Params, req *http.Request, render render.Render, chat *services.Chat) {
  tkn := req.Context().Value(conf.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    if !bson.IsObjectIdHex(params["id"]) {
      render.JSON(conf.ErrInvalidId.HttpCode, conf.ErrInvalidId)
      return
    }
    roomId := bson.ObjectIdHex(params["id"])
    if room, err := chat.GetRoom(profileID, roomId); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      render.JSON(http.StatusOK, room)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Deletes chat room
func DeleteRoom(params martini.Params, req *http.Request, res http.ResponseWriter, render render.Render, chat *services.Chat) {
  tkn := req.Context().Value(conf.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    if !bson.IsObjectIdHex(params["id"]) {
      render.JSON(conf.ErrInvalidId.HttpCode, conf.ErrInvalidId)
      return
    }
    roomId := bson.ObjectIdHex(params["id"])

    if err := chat.DeleteRoom(profileID, roomId); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      res.WriteHeader(http.StatusNoContent)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Adds member to the chat room
func AddRoomMember(params martini.Params, req *http.Request, res http.ResponseWriter, render render.Render, chat *services.Chat) {
  tkn := req.Context().Value(conf.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    if !bson.IsObjectIdHex(params["id"]) {
      render.JSON(conf.ErrInvalidId.HttpCode, conf.ErrInvalidId)
      return
    }
    roomId := bson.ObjectIdHex(params["id"])
    memberId := params["memberId"]

    if err := chat.AddRoomMember(profileID, roomId, memberId); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      res.WriteHeader(http.StatusOK)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Removes chat room member
func RemoveRoomMember(params martini.Params, req *http.Request, res http.ResponseWriter, render render.Render, chat *services.Chat) {
  tkn := req.Context().Value(conf.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    if !bson.IsObjectIdHex(params["id"]) {
      render.JSON(conf.ErrInvalidId.HttpCode, conf.ErrInvalidId)
      return
    }
    roomId := bson.ObjectIdHex(params["id"])
    memberId := params["memberId"]

    if err := chat.RemoveRoomMember(profileID, roomId, memberId); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      res.WriteHeader(http.StatusNoContent)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Send message to the chat log of a given room
func LogMessage(params martini.Params, message models.Message, req *http.Request, render render.Render, chat *services.Chat) {
  tkn := req.Context().Value(conf.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    if !bson.IsObjectIdHex(params["id"]) {
      render.JSON(conf.ErrInvalidId.HttpCode, conf.ErrInvalidId)
      return
    }
    roomId := bson.ObjectIdHex(params["id"])

    if newMessage, err := chat.LogMessage(profileID, roomId, message); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      render.JSON(http.StatusOK, newMessage)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Returns chat log of a given room
func GetChatLog(params martini.Params, req *http.Request, render render.Render, chat *services.Chat) {
  tkn := req.Context().Value(conf.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    if !bson.IsObjectIdHex(params["id"]) {
      render.JSON(conf.ErrInvalidId.HttpCode, conf.ErrInvalidId)
      return
    }
    roomId := bson.ObjectIdHex(params["id"])

    qs := req.URL.Query()
    fromParam, limitParam := qs.Get("from"), qs.Get("limit")

    var from time.Time
    if fromTS, err := strconv.ParseInt(fromParam, 10, 64); err != nil {
      from = time.Now() // If from is not a valid integer, ignore it
    } else {
      from = time.Unix(fromTS, 0)
    }
    limit, err := strconv.Atoi(limitParam)
    if err != nil {
      limit = conf.ChatLogLimit // If from is not a valid integer, ignore it
    }

    if messages, err := chat.GetChatLog(profileID, roomId, from, limit); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      render.JSON(http.StatusOK, messages)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}
