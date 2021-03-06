package controllers

import (
  "github.com/Askadias/go-chat-cluster/conf"
  "github.com/Askadias/go-chat-cluster/middleware"
  "github.com/Askadias/go-chat-cluster/models"
  "github.com/Askadias/go-chat-cluster/services"
  "github.com/codegangsta/martini-contrib/render"
  "github.com/dgrijalva/jwt-go"
  "github.com/go-martini/martini"
  "github.com/gorilla/websocket"
  "log"
  "net/http"
  "strconv"
  "time"
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
  deviceID middleware.DeviceID,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    if socket, err := upgrader.Upgrade(res, req, nil); err != nil {
      render.JSON(http.StatusInternalServerError, conf.NewApiError(err))
    } else {
      socket.SetReadLimit(conf.Socket.MaxMessageSize)
      socket.SetReadDeadline(time.Now().Add(conf.Socket.PongWait))
      socket.SetPongHandler(func(string) error {
        socket.SetReadDeadline(time.Now().Add(conf.Socket.PongWait))
        return nil
      })
      connection := &services.Connection{UserID: profileID, DeviceID: deviceID, Socket: socket}
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
  chatService *services.Chat,
  friendsService services.Friends,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)

    log.Println("Creating chat room for", profileID)

    if friends, err := friendsService.GetFriends(profileID); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      if !isFriendsOnly(room.Members, friends, profileID) {
        render.JSON(conf.ErrNotAFriend.HttpCode, conf.ErrNotAFriend)
        return
      }
    }

    if newRoom, err := chatService.CreateRoom(profileID, room); err != nil {
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
  chatService *services.Chat,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    if rooms, err := chatService.GetRooms(profileID); err != nil {
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
  chatService *services.Chat,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    roomID := params["id"]
    if room, err := chatService.GetRoom(profileID, roomID); err != nil {
      render.JSON(err.HttpCode, err)
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
  chatService *services.Chat,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    roomID := params["id"]

    if err := chatService.DeleteRoom(profileID, roomID); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
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
  render render.Render,
  chatService *services.Chat,
  friendsService services.Friends,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    roomID := params["id"]
    memberID := params["memberID"]

    if friends, err := friendsService.GetFriends(profileID); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      if !isFriendsOnly([]string{memberID}, friends, profileID) {
        render.JSON(conf.ErrNotAFriend.HttpCode, conf.ErrNotAFriend)
        return
      }
    }

    if room, err := chatService.AddRoomMember(profileID, roomID, memberID); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      render.JSON(http.StatusOK, room)
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
  chatService *services.Chat,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    roomID := params["id"]
    memberID := params["memberID"]

    if err := chatService.RemoveRoomMember(profileID, roomID, memberID); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
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
  chatService *services.Chat,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    roomID := params["id"]

    message.FromID = profileID
    message.RoomID = roomID
    if newMessage, err := chatService.AddMessage(message); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      render.JSON(http.StatusOK, newMessage)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Returns chat log of a given room
func GetChatLog(
  params martini.Params,
  req *http.Request,
  render render.Render,
  chatService *services.Chat,
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
    } else if limit > conf.Chat.MaxMessagesLimit {
      limit = conf.Chat.MaxMessagesLimit // upper limit
    }

    if messages, err := chatService.GetMessages(profileID, roomID, from, limit); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      render.JSON(http.StatusOK, messages)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Returns all members info for a given chat room
func GetAllMembersInfo(
  req *http.Request,
  params martini.Params,
  render render.Render,
  chatService *services.Chat,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    roomID := params["id"]
    if memberInfos, err := chatService.GetAllMembersInfo(profileID, roomID); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      render.JSON(http.StatusOK, memberInfos)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Returns all chat member infos for a given member
func GetAllRoomsInfo(
  req *http.Request,
  render render.Render,
  chatService *services.Chat,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    if memberInfos, err := chatService.GetAllRoomsInfo(profileID); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      render.JSON(http.StatusOK, memberInfos)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Returns room member info
func GetMemberInfo(
  params martini.Params,
  req *http.Request,
  render render.Render,
  chatService *services.Chat,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    roomID := params["id"]
    if memberInfo, err := chatService.GetMemberInfo(profileID, roomID); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      render.JSON(http.StatusOK, memberInfo)
    }
  } else {
    render.JSON(conf.ErrInvalidToken.HttpCode, conf.ErrInvalidToken)
  }
}

// Updates room messages last read time fo a given member
func UpdateMemberLastReadTime(
  params martini.Params,
  req *http.Request,
  res http.ResponseWriter,
  render render.Render,
  chatService *services.Chat,
) {
  tkn := req.Context().Value(conf.System.JWTUserPropName).(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"].(string)
    roomID := params["id"]
    if err := chatService.UpdateLastReadTime(profileID, roomID); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      res.WriteHeader(http.StatusNoContent)
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
