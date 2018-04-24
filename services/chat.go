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
  memberInfoDB      db.MemberInfo
  roomCache         db.RoomCache
  connectionManager *ConnectionManager
}

func NewChat(
  chatConf conf.ChatConf,
  chatDB db.Chat,
  chatLogDB db.ChatLog,
  memberInfoDB db.MemberInfo,
  roomCache db.RoomCache,
  connectionManager *ConnectionManager,
) *Chat {
  return &Chat{
    chatConf:          chatConf,
    chatDB:            chatDB,
    chatLogDB:         chatLogDB,
    memberInfoDB:      memberInfoDB,
    roomCache:         roomCache,
    connectionManager: connectionManager,
  }
}

// Checks if user exceeds the maximum of opened chat rooms and creates a new room
func (c *Chat) CreateRoom(profileID string, room models.Room) (*models.Room, *conf.ApiError) {
  // Check for max opened rooms
  if count, err := c.chatDB.OpenedRoomsCount(profileID); err != nil {
    return nil, conf.NewApiError(err)
    if count > c.chatConf.MaxOpenedRooms {
      return nil, conf.ErrTooManyChatsOpened
    }
  }
  // If new room is a group chat then set current user as an owner
  if len(room.Members) > 2 && room.OwnerID == "" {
    room.OwnerID = profileID
  }
  if room, err := c.chatDB.CreateRoom(room); err != nil {
    return nil, err
  } else {
    for _, memberID := range room.Members {
      memberInfo := models.MemberInfo{RoomID: room.ID, MemberID: memberID}
      if _, err := c.memberInfoDB.CreateMemberInfo(memberInfo); err != nil {
        log.Println("Failed to store member info for memberID:", memberID, "roomID:", room.ID, "error:", err.Message)
      }
    }
    c.roomCache.PutRoom(room.ID, room)
    return room, nil
  }
}

// Returns list of rooms that current user is member of
func (c *Chat) GetRooms(profileID string) ([]models.Room, *conf.ApiError) {
  if memberInfos, err := c.memberInfoDB.GetAllRoomsInfo(profileID); err != nil {
    return nil, err
  } else {
    roomIDs := make([]string, len(memberInfos))
    roomInfoIdx := make(map[string]models.MemberInfo)
    for i, memberInfo := range memberInfos {
      roomIDs[i] = memberInfo.RoomID
      roomInfoIdx[memberInfo.RoomID] = memberInfo
    }
    if rooms, err := c.chatDB.GetRoomsIn(roomIDs); err != nil {
      return nil, err
    } else {
      for _, room := range rooms {
        room.MemberInfo = roomInfoIdx[room.ID]
      }
      return rooms, nil
    }
  }
}

// Returns a specific room by id where current user is member of
func (c *Chat) GetRoom(profileID string, roomID string) (*models.Room, *conf.ApiError) {
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
    // 1. get all active room members to notify about the room membership update
    membersInfo, err := c.memberInfoDB.GetAllMembersInfo(roomID)
    if err != nil {
      return err
    }
    // 2. defer room membership update notification
    defer c.broadcastRoomUpdate(room, membersInfo)

    // 3. if user is an owner of the chat room or there are no any active members left in the room
    // then remove the room and cleanup its membership
    if room.OwnerID == profileID || len(membersInfo) <= 1 {
      if err := c.chatDB.DeleteRoom(roomID); err != nil {
        return err
      } else {
        c.roomCache.EvictRoom(roomID)
        for _, memberInfo := range membersInfo {
          if err := c.memberInfoDB.DeleteMemberInfo(roomID, memberInfo.MemberID); err != nil {
            log.Println("Unmable to delete member info roomID:", roomID, "memberID", memberInfo.MemberID)
          }
        }
        return nil
      }
    } else {
      // 4. if there are members left then just leave the room
      if err := c.memberInfoDB.DeleteMemberInfo(roomID, profileID); err != nil {
        return err
      }
    }

    // 5. private chat should kep history if there is at least one member
    if room.OwnerID == "" {
      return nil
    } else {
      // 6. if it's not a private chat but there are still members left then exclude member from room
      log.Println("Exiting chat roomID:", roomID, "memberID:", profileID)
      if err := c.chatDB.RemoveRoomMember(roomID, profileID); err != nil {
        return err
      } else {
        return nil
      }
    }
  }
}

func (c *Chat) broadcastRoomUpdate(room *models.Room, membersInfo []models.MemberInfo) {
  memberIDs := make([]string, len(membersInfo))
  for i, memberInfo := range membersInfo {
    memberIDs[i] = memberInfo.MemberID
  }
  c.connectionManager.Broadcast <- &BroadcastPackage{
    Message:  &models.Message{Type: "update", RoomID: room.ID},
    Auditory: memberIDs,
  }
}

func (c *Chat) broadcastMessage(room *models.Room, message *models.Message) {
  membersInfo, err := c.memberInfoDB.GetAllMembersInfo(room.ID)
  if err != nil {
    log.Println("Can't get members info roomID:", room.ID, "err:", err)
    return
  }
  memberIDs := make([]string, len(membersInfo))
  roomInfoIdx := make(map[string]models.MemberInfo)
  for i, memberInfo := range membersInfo {
    memberIDs[i] = memberInfo.MemberID
    roomInfoIdx[memberInfo.RoomID] = memberInfo
  }
  for _, memberID := range room.Members {
    if _, ok := roomInfoIdx[memberID]; !ok {
      newMemberInfo := models.MemberInfo{RoomID: room.ID, MemberID: memberID}
      if _, err := c.memberInfoDB.CreateMemberInfo(newMemberInfo); err != nil {
        log.Println("Cannot create member info roomID:", room.ID, "memberID:", memberID, "err:", err)
      }
    }
  }
  c.connectionManager.Broadcast <- &BroadcastPackage{Message: message, Auditory: room.Members}
}

// Add room member to a specific chat room.
// Checks if members count exceeds the configured maximum
func (c *Chat) AddRoomMember(profileID string, roomID string, memberID string) (*models.Room, *conf.ApiError) {
  if room, err := c.GetRoom(profileID, roomID); err != nil {
    return nil, err
  } else {
    if isMember(memberID, room) {
      return nil, conf.ErrAlreadyExists
    }
    room.Members = append(room.Members, memberID)
    // if current room is a private one, then create a new room for the group chat
    if len(room.Members) == 3 && room.OwnerID == "" {
      if newRoom, err := c.CreateRoom(profileID, *room); err != nil {
        return nil, err
      } else {
        return newRoom, nil
      }
    }
    if len(room.Members) >= c.chatConf.MaxMembers {
      return nil, conf.ErrTooManyMembers
    }
    if err := c.chatDB.AddRoomMember(roomID, memberID); err != nil {
      return nil, err
    }

    membersInfo, err := c.memberInfoDB.GetAllMembersInfo(roomID)
    if err != nil {
      return nil, err
    }
    memberInfo := models.MemberInfo{RoomID: room.ID, MemberID: memberID}
    if _, err := c.memberInfoDB.CreateMemberInfo(memberInfo); err != nil {
      log.Println("Failed to store member info for memberID:", memberID, "roomID:", room.ID, "error:", err.Message)
    }
    c.roomCache.PutRoom(roomID, room) // TODO concurrency check
    membersInfo = append(membersInfo, memberInfo)
    defer c.broadcastRoomUpdate(room, membersInfo)
    return room, nil
  }
}

func isMember(userID string, room *models.Room) bool {
  for _, memberID := range room.Members {
    if memberID == userID {
      return true
    }
  }
  return false
}

// Removes room member by the room owner.
// If there are no more members then delete the room entirely.
func (c *Chat) RemoveRoomMember(profileID string, roomID string, memberID string) *conf.ApiError {
  if room, err := c.GetRoom(profileID, roomID); err != nil {
    return err
  } else {
    if room.OwnerID != profileID {
      return conf.ErrNotAnOwner
    }
    if err := c.chatDB.RemoveRoomMember(roomID, memberID); err != nil {
      return err
    }
    membersInfo, err := c.memberInfoDB.GetAllMembersInfo(roomID)
    if err != nil {
      return err
    }
    defer c.broadcastRoomUpdate(room, membersInfo)
    if err := c.memberInfoDB.DeleteMemberInfo(roomID, memberID); err != nil {
      log.Println("Failed to remove member info for memberID:", memberID, "roomID:", room.ID, "error:", err.Message)
    }
    room.Updated = models.Timestamp(time.Now())
    if len(room.Members) <= 2 {
      if err := c.chatDB.DeleteRoom(roomID); err != nil {
        return err
      } else {
        c.roomCache.EvictRoom(roomID)
        if err := c.memberInfoDB.DeleteAllMembersInfo(roomID); err != nil {
          log.Println("Failed to cleanup member info for roomID:", room.ID, "error:", err.Message)
        }
      }
    } else {
      c.roomCache.EvictRoom(roomID)
      //c.roomCache.PutRoom(roomID, room) TODO verify cache updates properly
    }
    return nil
  }
}

// Adds new message to the chat log
// Additionally checks that current user is a member of that chat room.
func (c *Chat) AddMessage(message models.Message) (*models.Message, *conf.ApiError) {
  if err := c.chatDB.IsRoomMember(message.RoomID, message.FromID); err != nil {
    return nil, err
  }
  if message, err := c.chatLogDB.AddMessage(message); err != nil {
    return nil, err
  } else {
    if room, err := c.GetRoom(message.FromID, message.RoomID); err != nil {
      return nil, err
    } else {
      defer c.broadcastMessage(room, message)
      return message, nil
    }
  }
}

// Returns chat log for a given period
func (c *Chat) GetMessages(profileID string, roomID string, from time.Time, limit int) ([]models.Message, *conf.ApiError) {
  if err := c.chatDB.IsRoomMember(roomID, profileID); err != nil {
    return nil, err
  }
  return c.chatLogDB.GetMessages(profileID, roomID, from, limit)
}

// Adds reaction to a specific message
func (c *Chat) AddReaction(profileID string, roomID string, messageID string, reaction string) *conf.ApiError {
  if err := c.chatDB.IsRoomMember(roomID, profileID); err != nil {
    return err
  }
  return c.chatLogDB.AddReaction(profileID, messageID, reaction)
}

// Modifies the specific message
func (c *Chat) EditMessage(profileID string, messageID string, body string) *conf.ApiError {
  return c.chatLogDB.EditMessage(profileID, messageID, body)
}

// Updates last read time for a given chat member
func (c *Chat) UpdateLastReadTime(profileID string, roomID string) *conf.ApiError {
  if err := c.memberInfoDB.UpdateLastReadTime(roomID, profileID); err != nil {
    log.Println("Failed to update member info for memberID:", profileID, "roomID:", roomID, "error:", err.Message)
    return err
  } else {
    return nil
  }
}

// Returns all chat member infos for a given member
func (c *Chat) GetAllRoomsInfo(profileID string) ([]models.MemberInfo, *conf.ApiError) {
  return c.memberInfoDB.GetAllRoomsInfo(profileID)
}

// Returns all chat members info
func (c *Chat) GetAllMembersInfo(profileID string, roomID string) ([]models.MemberInfo, *conf.ApiError) {
  if err := c.chatDB.IsRoomMember(roomID, profileID); err != nil {
    return nil, err
  }
  return c.memberInfoDB.GetAllMembersInfo(roomID)
}

// Returns chat member info
func (c *Chat) GetMemberInfo(profileID string, roomID string) (*models.MemberInfo, *conf.ApiError) {
  if err := c.chatDB.IsRoomMember(roomID, profileID); err != nil {
    return nil, err
  }
  return c.memberInfoDB.GetMemberInfo(roomID, profileID)
}
