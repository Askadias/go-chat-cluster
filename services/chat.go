package services

import (
  "github.com/Askadias/go-chat-cluster/db"
  "github.com/Askadias/go-chat-cluster/models"
  "github.com/Askadias/go-chat-cluster/conf"
  "log"
  "time"
)

type Chat struct {
  chatConf          conf.ChatConf
  chatDB            db.Chat
  chatLogDB         db.ChatLog
  roomCache         db.RoomCache
  connectionManager *ConnectionManager
}

func NewChat(
  chatConf conf.ChatConf,
  chatDB db.Chat,
  chatLogDB db.ChatLog,
  roomCache db.RoomCache,
  connectionManager *ConnectionManager,
) *Chat {
  return &Chat{
    chatConf:          chatConf,
    chatDB:            chatDB,
    chatLogDB:         chatLogDB,
    roomCache:         roomCache,
    connectionManager: connectionManager,
  }
}

// Checks if user exceeds the maximum of opened chat rooms and creates a new room
func (c *Chat) CreateRoom(profileID string, room models.Room) (*models.Room, *conf.ApiError) {
  if count, err := c.chatDB.OpenedRoomsCount(profileID); err != nil {
    return nil, conf.NewApiError(err)
    if count > c.chatConf.MaxOpenedRooms {
      return nil, conf.ErrTooManyChatsOpened
    }
  }
  log.Println("Creating new chat room for", profileID)
  return c.chatDB.CreateRoom(profileID, room)
}

// Returns list of rooms that current user is member of
func (c *Chat) GetRooms(profileID string) ([]models.Room, *conf.ApiError) {
  log.Println("Retrieving chat rooms for", profileID)
  return c.chatDB.GetRooms(profileID)
}

// Returns a specific room by id where current user is member of
func (c *Chat) GetRoom(profileID string, roomID string) (*models.Room, *conf.ApiError) {
  log.Println("Retrieving chat room", roomID, "by", profileID)
  if room, err := c.roomCache.GetRoom(roomID); err != nil {
    if room, err := c.chatDB.GetRoom(profileID, roomID); err != nil {
      return nil, err
    } else {
      c.roomCache.PutRoom(roomID, room)
      return room, nil
    }
  } else {
    return room, nil
  }
}

// Remove room association of the current user.
// If user is an owner of the specific group chat then this chat will be deleted.
func (c *Chat) DeleteRoom(profileID string, roomID string) *conf.ApiError {
  if room, err := c.GetRoom(profileID, roomID); err != nil {
    return err
  } else {
    // if not an owner then exit this room
    if room.Owner != profileID {
      log.Println("Exiting chat room", roomID, "by member", profileID)
      if err := c.chatDB.RemoveRoomMember(roomID, profileID); err != nil {
        return err
      } else {
        return nil
      }
    }
    log.Println("Deleting chat room", roomID, "by owner", profileID)
    if err := c.chatDB.DeleteRoom(roomID); err != nil {
      return err
    } else {
      c.roomCache.EvictRoom(roomID)
      c.connectionManager.Broadcast <- &BroadcastPackage{
        Message:  &models.Message{Type: "update", Room: roomID},
        Auditory: room.Members,
      }
      return nil
    }
  }
}

// Add room member to a specific chat room.
// Checks if members count exceeds the configured maximum
func (c *Chat) AddRoomMember(profileID string, roomID string, memberID string) *conf.ApiError {
  if room, err := c.GetRoom(profileID, roomID); err != nil {
    return err
  } else {
    if len(room.Members) >= c.chatConf.MaxMembers {
      return conf.ErrTooManyMembers
    }
    if err := c.chatDB.AddRoomMember(roomID, memberID); err != nil {
      return err
    }
    c.roomCache.PutRoom(roomID, room) // TODO concurrency check
    c.connectionManager.Broadcast <- &BroadcastPackage{
      Message:  &models.Message{Type: "update", Room: roomID},
      Auditory: append(room.Members, memberID),
    }
    return nil
  }
}

// Removes room member by the room owner.
// If there are no more members then delete the room entirely.
func (c *Chat) RemoveRoomMember(profileID string, roomID string, memberID string) *conf.ApiError {
  if room, err := c.GetRoom(profileID, roomID); err != nil {
    return err
  } else {
    if room.Owner != profileID {
      return conf.ErrNotAnOwner
    }
    log.Println("Remove member", memberID, "from room", roomID, "by", profileID)
    if err := c.chatDB.RemoveRoomMember(roomID, memberID); err != nil {
      return err
    }
    room.Updated = models.Timestamp(time.Now())
    if len(room.Members) <= 2 {
      if err := c.chatDB.DeleteRoom(roomID); err != nil {
        return err
      } else {
        c.roomCache.EvictRoom(roomID)
      }
    } else {
      c.roomCache.PutRoom(roomID, room)
    }
    c.connectionManager.Broadcast <- &BroadcastPackage{
      Message:  &models.Message{Type: "update", Room: roomID},
      Auditory: append(room.Members, memberID),
    }
    return nil
  }
}

// Adds new message to the chat log
// Additionally checks that current user is a member of that chat room.
func (c *Chat) AddMessage(message models.Message) (*models.Message, *conf.ApiError) {
  if err := c.chatDB.IsRoomMember(message.Room, message.From); err != nil {
    return nil, err
  }
  log.Println("Log message from", message.From, "to room", message.Room)
  if message, err := c.chatLogDB.AddMessage(message); err != nil {
    return nil, err
  } else {
    if room, err := c.GetRoom(message.From, message.Room); err != nil {
      return nil, err
    } else {
      c.connectionManager.Broadcast <- &BroadcastPackage{Message: message, Auditory: room.Members}
      return message, nil
    }
  }
}

// Returns chat log for a given period
func (c *Chat) GetMessages(profileID string, roomID string, from time.Time, limit int) ([]models.Message, *conf.ApiError) {
  if err := c.chatDB.IsRoomMember(roomID, profileID); err != nil {
    return nil, err
  }
  log.Println("Retrieving chat log from:", from, "limit:", limit, "room:", roomID, "member:", profileID)
  return c.chatLogDB.GetMessages(profileID, roomID, from, limit)
}

// Adds reaction to a specific message
func (c *Chat) AddReaction(profileID string, roomID string, messageID string, reaction string) *conf.ApiError {
  if err := c.chatDB.IsRoomMember(roomID, profileID); err != nil {
    return err
  }
  log.Println("Adding reacion", reaction, "to a message", messageID, "by", profileID)
  return c.chatLogDB.AddReaction(profileID, messageID, reaction)
}

// Modifies the specific message
func (c *Chat) EditMessage(profileID string, messageID string, body string) *conf.ApiError {
  log.Println("Editing message:", messageID, "by", profileID)
  return c.chatLogDB.EditMessage(profileID, messageID, body)
}
