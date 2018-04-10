package controllers

import (
  "net/http"
  "github.com/gorilla/websocket"
  "github.com/Askadias/go-chat-cluster/services"
  "github.com/dgrijalva/jwt-go"
  "github.com/codegangsta/martini-contrib/render"
  "github.com/Askadias/go-chat-cluster/conf"
  "log"
  "github.com/Askadias/go-chat-cluster/models"
  "time"
  "github.com/go-martini/martini"
  "strconv"
  "github.com/Askadias/go-chat-cluster/db"
)

var upgrader = websocket.Upgrader{
  HandshakeTimeout: conf.Socket.HandshakeTimeout,
  ReadBufferSize:   conf.Socket.ReadBufferSize,
  WriteBufferSize:  conf.Socket.WriteBufferSize,
}

// Initializes a WebSocket connection for the current user
func ConnectToChat(
  req *http.Request,
  res http.ResponseWriter,
  render render.Render,
  manager *services.ConnectionManager,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    if conn, err := upgrader.Upgrade(res, req, nil); err != nil {
      render.JSON(http.StatusInternalServerError, conf.NewApiError(err))
    } else {
      conn.SetReadLimit(conf.Socket.MaxMessageSize)
      conn.SetReadDeadline(time.Now().Add(conf.Socket.PongWait))
      conn.SetPongHandler(func(string) error {
        conn.SetReadDeadline(time.Now().Add(conf.Socket.PongWait))
        return nil
      })
      connection := services.NewConnection(profileID, conn)
      manager.Register <- connection
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Create chat room
func CreateRoom(
  room models.Room,
  req *http.Request,
  render render.Render,
  chat db.Chat,
  friends services.Friends,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)

    log.Println("Creating chat room for", profileID)

    if friends, err := friends.GetFriends(profileID); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      if !isFriendsOnly(room.Members, friends, profileID) {
        render.JSON(conf.ErrNotAFriend.HttpCode, conf.ErrNotAFriend)
        return
      }
    }

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
func GetRooms(
  req *http.Request,
  render render.Render,
  chat db.Chat,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
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
func GetRoom(
  params martini.Params,
  req *http.Request,
  render render.Render,
  chat db.Chat,
  roomCache db.RoomCache,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    roomID := params["id"]
    if room, err := roomCache.GetRoom(roomID); err != nil {
      if room, err := chat.GetRoom(profileID, roomID); err != nil {
        render.JSON(err.HttpCode, err)
      } else {
        render.JSON(http.StatusOK, room)
        roomCache.PutRoom(roomID, room)
      }
    } else {
      render.JSON(http.StatusOK, room)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Deletes chat room
func DeleteRoom(
  params martini.Params,
  req *http.Request,
  res http.ResponseWriter,
  render render.Render,
  chat db.Chat,
  manager *services.ConnectionManager,
  roomCache db.RoomCache,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    roomID := params["id"]

    if room, err := chat.DeleteRoom(profileID, roomID); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      roomCache.EvictRoom(roomID)
      manager.Broadcast <- &services.BroadcastPackage{
        Message:  &models.Message{Type: "update", Room: roomID},
        Auditory: room.Members,
      }
      res.WriteHeader(http.StatusNoContent)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Adds member to the chat room
func AddRoomMember(
  params martini.Params,
  req *http.Request,
  res http.ResponseWriter,
  render render.Render,
  chat db.Chat,
  friends services.Friends,
  manager *services.ConnectionManager,
  roomCache db.RoomCache,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    roomID := params["id"]
    memberID := params["memberID"]

    if friends, err := friends.GetFriends(profileID); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      if !isFriendsOnly([]string{memberID}, friends, profileID) {
        render.JSON(conf.ErrNotAFriend.HttpCode, conf.ErrNotAFriend)
        return
      }
    }

    if room, err := chat.AddRoomMember(profileID, roomID, memberID); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      roomCache.PutRoom(roomID, room)
      manager.Broadcast <- &services.BroadcastPackage{
        Message:  &models.Message{Type: "update", Room: roomID},
        Auditory: room.Members,
      }
      res.WriteHeader(http.StatusOK)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Removes chat room member
func RemoveRoomMember(
  params martini.Params,
  req *http.Request,
  res http.ResponseWriter,
  render render.Render,
  chat db.Chat,
  manager *services.ConnectionManager,
  roomCache db.RoomCache,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    roomID := params["id"]
    memberID := params["memberID"]

    if room, err := chat.RemoveRoomMember(profileID, roomID, memberID); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      if len(room.Members) <= 2 {
        roomCache.EvictRoom(roomID) // room was deleted
      } else {
        roomCache.PutRoom(roomID, room)
      }
      manager.Broadcast <- &services.BroadcastPackage{
        Message:  &models.Message{Type: "update", Room: roomID},
        Auditory: append(room.Members, memberID),
      }
      res.WriteHeader(http.StatusNoContent)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Send message to the chat log of a given room
func SendMessage(
  params martini.Params,
  message models.Message,
  req *http.Request,
  render render.Render,
  chat db.Chat,
  chatLog db.ChatLog,
  manager *services.ConnectionManager,
  roomCache db.RoomCache,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    roomID := params["id"]

    message.From = profileID
    message.Room = roomID
    if newMessage, err := chatLog.AddMessage(message); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      if room, err := roomCache.GetRoom(roomID); err != nil {
        if room, err := chat.GetRoom(profileID, roomID); err != nil {
          render.JSON(err.HttpCode, err)
        } else {
          broadcast(newMessage, room.Members, manager, render)
          roomCache.PutRoom(roomID, room)
        }
      } else {
        broadcast(newMessage, room.Members, manager, render)
      }
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

func broadcast(message *models.Message, auditory []string, manager *services.ConnectionManager, render render.Render) {
  manager.Broadcast <- &services.BroadcastPackage{Message: message, Auditory: auditory}
  render.JSON(http.StatusOK, message)
}

// Returns chat log of a given room
func GetChatLog(
  params martini.Params,
  req *http.Request,
  render render.Render,
  chatLog db.ChatLog,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    roomID := params["id"]

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
      limit = conf.Chat.DefaultMessagesLimit // If from is not a valid integer, ignore it
    }

    if messages, err := chatLog.GetMessages(profileID, roomID, from, limit); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      render.JSON(http.StatusOK, messages)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

func isFriendsOnly(members []string, friends []models.User, selfID string) bool {
  allowed := make(map[string]struct{})
  allowed[selfID] = struct{}{}
  for _, friend := range friends {
    allowed[friend.ID] = struct{}{}
  }
  for _, m := range members {
    if _, ok := allowed[m]; !ok {
      return false
    }
  }
  return true
}
