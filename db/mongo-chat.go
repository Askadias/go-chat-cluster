package db

import (
  "gopkg.in/mgo.v2"
  "crypto/tls"
  "net"
  "log"
  "time"
  "github.com/Askadias/go-chat-cluster/models"
  "github.com/Askadias/go-chat-cluster/conf"
  "gopkg.in/mgo.v2/bson"
  "strings"
)

// Chat Service manages chat rooms and their members.
type MongoChat struct {
  mongoSession *mgo.Session
  mongo        conf.MongoConf
}

// Creates Chat Service. Initialises MongoDB connection.
func NewMongoChat(mongoOptions conf.MongoConf) *MongoChat {
  dialInfo, err := mgo.ParseURL(mongoOptions.URL)

  tlsConfig := &tls.Config{}
  dialInfo.Timeout = mongoOptions.Timeout
  dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
    conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
    return conn, err
  }
  session, err := mgo.DialWithInfo(dialInfo)
  if err != nil {
    log.Fatalf("CreateSession: %s\n", err)
  }
  return &MongoChat{
    mongoSession: session,
    mongo:        mongoOptions,
  }
}

// Returns number of opened rooms for a given member
func (c *MongoChat) OpenedRoomsCount(memberID string) (int, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if count, err := db.C(c.mongo.RoomCollectionName).Find(bson.M{"owner": memberID}).Count(); err != nil {
    return -1, parseMongoDBError(err)
  } else {
    return count, nil
  }
}

// Save new RoomID into the MongoDB by setting current profile ID as an owner.
// Automatically adds current user to the chat room members.
func (c *MongoChat) CreateRoom(room models.Room) (*models.Room, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  room.ID = bson.NewObjectId().Hex()
  now := models.Timestamp(time.Now())
  room.Created = now
  room.Updated = now

  if err := db.C(c.mongo.RoomCollectionName).Insert(room); err != nil {
    return nil, parseMongoDBError(err)
  }
  return &room, nil
}

// Returns all the ChatRooms by a given member
func (c *MongoChat) GetRooms(memberID string) ([]models.Room, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  var rooms []models.Room
  if err := db.C(c.mongo.RoomCollectionName).Find(bson.M{"members": memberID}).All(&rooms); err != nil {
    return nil, conf.NewApiError(err)
  }
  return rooms, nil
}

// Returns all the ChatRooms by a given list of IDs
func (c *MongoChat) GetRoomsIn(roomIDs []string) ([]models.Room, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  var rooms []models.Room
  if err := db.C(c.mongo.RoomCollectionName).Find(bson.M{"_id": bson.M{"$in": roomIDs}}).All(&rooms); err != nil {
    return nil, conf.NewApiError(err)
  }
  return rooms, nil
}

// Returns RoomID by ID and membership
func (c *MongoChat) GetRoom(memberID string, roomID string) (*models.Room, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if !bson.IsObjectIdHex(roomID) {
    return nil, conf.ErrInvalidId
  }

  var room models.Room
  if err := db.C(c.mongo.RoomCollectionName).Find(bson.M{"_id": roomID, "members": memberID}).One(&room); err != nil {
    return nil, parseMongoDBError(err)
  }
  return &room, nil
}

// Deletes RoomID from the MongoDB by its owner.
func (c *MongoChat) DeleteRoom(roomID string) *conf.ApiError {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if err := db.C(c.mongo.RoomCollectionName).RemoveId(roomID); err != nil {
    return parseMongoDBError(err)
  }
  return nil
}

// Adds new member to a given chat room.
// Actual member info retrieved from account service.
// Additionally checks that current user is an owner of that room to update it.
func (c *MongoChat) AddRoomMember(roomID string, memberID string) *conf.ApiError {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  now := models.Timestamp(time.Now())
  if err := db.C(c.mongo.RoomCollectionName).Update(
    bson.M{"_id": roomID, "members": bson.M{"$ne": memberID}},
    bson.M{
      "$push": bson.M{"members": memberID},
      "$set": bson.M{
        "updated":                    now,
        "memberAdded." + memberID:    now,
        "memberLastRead." + memberID: now,
      },
    }); err != nil {
    return parseMongoDBError(err)
  }
  return nil
}

// Adds new member to a given chat room.
// Actual member info retrieved from account service.
// Additionally checks that current user is an owner of that room to update it.
func (c *MongoChat) TouchRoomMember(roomID string, memberID string) *conf.ApiError {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  now := models.Timestamp(time.Now())
  if err := db.C(c.mongo.RoomCollectionName).Update(
    bson.M{"_id": roomID, "members": memberID},
    bson.M{
      "$set": bson.M{
        "updated":                    now,
        "memberLastRead." + memberID: now,
      },
    }); err != nil {
    return parseMongoDBError(err)
  }
  return nil
}

// Adds new member to a given chat room.
// Actual member info retrieved from account service.
// Additionally checks that current user is an owner of that room to update it.
func (c *MongoChat) UpdateLastMessageRoomMember(roomID string, memberID string) *conf.ApiError {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  now := models.Timestamp(time.Now())
  if err := db.C(c.mongo.RoomCollectionName).Update(
    bson.M{"_id": roomID, "members": memberID},
    bson.M{
      "$set": bson.M{
        "updated":                    now,
        "memberLastRead." + memberID: now,
      },
    }); err != nil {
    return parseMongoDBError(err)
  }
  return nil
}

// Updates current RoomID by removing a member with a specific id.
// Additionally checks that current user is an owner of that room to update it.
func (c *MongoChat) RemoveRoomMember(roomID string, memberID string) *conf.ApiError {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if err := db.C(c.mongo.RoomCollectionName).Update(
    bson.M{"_id": roomID, "members": memberID},
    bson.M{
      "$pull": bson.M{"members": memberID},
      "$set":  bson.M{"updated": time.Now()},
      "$unset": bson.M{
        "memberAdded." + memberID:    "",
        "memberLastRead." + memberID: "",
      },
    }); err != nil {
    return parseMongoDBError(err)
  }
  return nil
}

func (c *MongoChat) IsRoomMember(roomID string, memberID string) *conf.ApiError {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if !bson.IsObjectIdHex(roomID) {
    return conf.ErrInvalidId
  }

  if count, err := db.C(c.mongo.RoomCollectionName).Find(bson.M{"_id": roomID, "members": memberID}).Count(); err != nil {
    return parseMongoDBError(err)
  } else if count == 0 {
    return conf.ErrNotAMember
  }
  return nil
}

// Adds message to chat log to a given chat room by a given member.
func (c *MongoChat) AddMessage(message models.Message) (*models.Message, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if !bson.IsObjectIdHex(message.RoomID) {
    return nil, conf.ErrInvalidId
  }

  message.ID = bson.NewObjectId().Hex()
  message.Reactions = make(map[string]string)
  message.Timestamp = models.Timestamp(time.Now())

  if err := db.C(c.mongo.MessagesCollectionName).Insert(message); err != nil {
    return nil, parseMongoDBError(err)
  }
  return &message, nil
}

// Returns chat log page by a given start timestamp and limit.
// Additionally checks that current user is a member of that chat room.
func (c *MongoChat) GetMessages(profileID string, roomID string, from time.Time, limit int) ([]models.Message, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if !bson.IsObjectIdHex(roomID) {
    return nil, conf.ErrInvalidId
  }

  var messages []models.Message
  if err := db.C(c.mongo.MessagesCollectionName).Find(
    bson.M{"room": roomID, "timestamp": bson.M{"$lt": from}}).Limit(limit).Sort("-timestamp").All(&messages); err != nil {
    return nil, parseMongoDBError(err)
  }
  return messages, nil
}

func (c *MongoChat) AddReaction(profileID string, messageID string, reaction string) *conf.ApiError {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if err := db.C(c.mongo.MessagesCollectionName).UpdateId(messageID,
    bson.M{"$set": bson.M{"reaction." + profileID: reaction}}); err != nil {
    return parseMongoDBError(err)
  } else {
    return nil
  }
}

func (c *MongoChat) EditMessage(profileID string, messageID string, body string) *conf.ApiError {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if err := db.C(c.mongo.MessagesCollectionName).Update(
    bson.M{"_id": messageID, "from": profileID},
    bson.M{"$set": bson.M{"body": body}}); err != nil {
    return parseMongoDBError(err)
  } else {
    return nil
  }
}

// Creates Member Info and returns last version
func (c *MongoChat) CreateMemberInfo(memberInfo models.MemberInfo) (*models.MemberInfo, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if !bson.IsObjectIdHex(memberInfo.RoomID) {
    return nil, conf.ErrInvalidId
  }

  now := models.Timestamp(time.Now())
  memberInfo.ID = memberInfo.RoomID + "|" + memberInfo.MemberID
  memberInfo.JoinedAt = now
  memberInfo.LastReadAt = now

  if err := db.C(c.mongo.MemberInfoCollectionName).Insert(memberInfo); err != nil {
    return nil, parseMongoDBError(err)
  }
  return &memberInfo, nil
}

// Returns current member info by id
func (c *MongoChat) GetMemberInfo(roomID string, memberID string) (*models.MemberInfo, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if !bson.IsObjectIdHex(roomID) {
    return nil, conf.ErrInvalidId
  }

  var memberInfo models.MemberInfo
  if err := db.C(c.mongo.MemberInfoCollectionName).FindId(roomID + "|" + memberID).One(&memberInfo); err != nil {
    return nil, parseMongoDBError(err)
  }
  return &memberInfo, nil
}

// Returns all the members metadata by a given roomID
func (c *MongoChat) GetAllMembersInfo(roomID string) ([]models.MemberInfo, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  var membersInfo []models.MemberInfo
  if err := db.C(c.mongo.MemberInfoCollectionName).Find(bson.M{"room": roomID}).All(&membersInfo); err != nil {
    return nil, conf.NewApiError(err)
  }
  return membersInfo, nil
}

// Returns all the rooms member metadata by a given memberID
func (c *MongoChat) GetAllRoomsInfo(memberID string) ([]models.MemberInfo, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  var membersInfo []models.MemberInfo
  if err := db.C(c.mongo.MemberInfoCollectionName).Find(bson.M{"member": memberID}).All(&membersInfo); err != nil {
    return nil, conf.NewApiError(err)
  }
  return membersInfo, nil
}

// Updates Member Info last read time
func (c *MongoChat) UpdateLastReadTime(roomID string, memberID string) *conf.ApiError {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if !bson.IsObjectIdHex(roomID) {
    return conf.ErrInvalidId
  }

  if err := db.C(c.mongo.MemberInfoCollectionName).UpdateId(
    roomID + "|" + memberID,
    bson.M{"$set": bson.M{"lastReadAt":  models.Timestamp(time.Now())}}); err != nil {
    return parseMongoDBError(err)
  } else {
    return nil
  }
}

// Deletes Member Info by id
func (c *MongoChat) DeleteMemberInfo(roomID string, memberID string) *conf.ApiError {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if !bson.IsObjectIdHex(roomID) {
    return conf.ErrInvalidId
  }

  if err := db.C(c.mongo.MemberInfoCollectionName).RemoveId(roomID + "|" + memberID); err != nil {
    return parseMongoDBError(err)
  }
  return nil
}

// Cleanup Members Info by roomID
func (c *MongoChat) DeleteAllMembersInfo(roomID string) *conf.ApiError {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if !bson.IsObjectIdHex(roomID) {
    return conf.ErrInvalidId
  }

  if _, err := db.C(c.mongo.MemberInfoCollectionName).RemoveAll(bson.M{"room": roomID}); err != nil {
    return parseMongoDBError(err)
  }
  return nil
}

func parseMongoDBError(err error) *conf.ApiError {
  if err.Error() == "not found" {
    return conf.ErrNotFound
  } else if strings.Contains(err.Error(), "duplicate key error") {
    return conf.ErrAlreadyExists
  } else {
    return conf.NewApiError(err)
  }
}
