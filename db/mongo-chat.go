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
  chat         conf.ChatConf
}

// Creates Chat Service. Initialises MongoDB connection.
func NewMongoChat(mongoOptions conf.MongoConf, chatOptions conf.ChatConf) *MongoChat {
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
    chat:         chatOptions,
  }
}

// Save new Room into the MongoDB by setting current profile ID as an owner.
// Automatically adds current user to the chat room members.
func (c *MongoChat) CreateRoom(profileID string, room models.Room) (*models.Room, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if count, _ := db.C(c.mongo.RoomCollectionName).Find(bson.M{"owner": profileID}).Count();
    count > c.chat.MaxOpenedRooms {
    return nil, conf.ErrTooManyChatsOpened
  }

  room.Owner = profileID
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
func (c *MongoChat) GetRooms(profileID string) ([]models.Room, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  log.Println("Retrieving chat rooms for", profileID)

  var rooms []models.Room
  if err := db.C(c.mongo.RoomCollectionName).Find(bson.M{"members": profileID}).All(&rooms); err != nil {
    return nil, conf.NewApiError(err)
  }
  return rooms, nil
}

// Returns Room by ID and membership
func (c *MongoChat) GetRoom(profileID string, roomID string) (*models.Room, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if !bson.IsObjectIdHex(roomID) {
    return nil, conf.ErrInvalidId
  }

  log.Println("Retrieving chat room", roomID, "by", profileID)

  var room models.Room
  if err := db.C(c.mongo.RoomCollectionName).Find(bson.M{"_id": roomID, "members": profileID}).One(&room); err != nil {
    return nil, parseMongoDBError(err)
  }
  return &room, nil
}

// Deletes Room from the MongoDB by its owner.
func (c *MongoChat) DeleteRoom(profileID string, roomID string) (*models.Room, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if !bson.IsObjectIdHex(roomID) {
    return nil, conf.ErrInvalidId
  }

  var room models.Room
  if err := db.C(c.mongo.RoomCollectionName).FindId(roomID).One(&room); err != nil {
    return nil, parseMongoDBError(err)
  }
  // if not an owner then exit this room
  if room.Owner != profileID {
    if err := c.doRemoveMember(db, room, profileID); err != nil {
      return nil, err
    } else {
      return &room, nil
    }
  }

  log.Println("Deleting chat room", roomID, "by owner", profileID)
  if err := db.C(c.mongo.RoomCollectionName).RemoveId(roomID); err != nil {
    return nil, parseMongoDBError(err)
  }
  return &room, nil
}

// Adds new member to a given chat room.
// Actual member info retrieved from account service.
// Additionally checks that current user is an owner of that room to update it.
func (c *MongoChat) AddRoomMember(profileID string, roomID string, memberID string) (*models.Room, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if !bson.IsObjectIdHex(roomID) {
    return nil, conf.ErrInvalidId
  }

  var room models.Room
  if err := db.C(c.mongo.RoomCollectionName).FindId(roomID).One(&room); err != nil {
    return nil, parseMongoDBError(err)
  }

  if len(room.Members) >= c.chat.MaxMembers {
    return nil, conf.ErrTooManyMembers
  }

  log.Println("Add member", memberID, "to room", roomID, "by", profileID)
  now := time.Now()

  if err := db.C(c.mongo.RoomCollectionName).Update(
    bson.M{"_id": roomID, "members": bson.M{"$ne": memberID}},
    bson.M{
      "$push": bson.M{"members": memberID},
      "$set":  bson.M{"updated": now},
    }); err != nil {
    return nil, parseMongoDBError(err)
  }
  room.Members = append(room.Members, memberID)
  room.Updated = models.Timestamp(now)
  return &room, nil
}

// Updates current Room by removing a member with a specific id.
// Additionally checks that current user is an owner of that room to update it.
func (c *MongoChat) RemoveRoomMember(profileID string, roomID string, memberID string) (*models.Room, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if !bson.IsObjectIdHex(roomID) {
    return nil, conf.ErrInvalidId
  }

  var room models.Room
  if err := db.C(c.mongo.RoomCollectionName).FindId(roomID).One(&room); err != nil {
    return nil, parseMongoDBError(err)
  }
  if room.Owner != profileID {
    return nil, conf.ErrNotAnOwner
  }

  log.Println("Remove member", memberID, "from room", roomID, "by", profileID)

  if err := c.doRemoveMember(db, room, memberID); err != nil {
    return nil, err
  }

  for i, m := range room.Members {
    if m == memberID {
      room.Members = append(room.Members[:i], room.Members[i+1:]...)
      break
    }
  }
  return &room, nil
}

func (c *MongoChat) doRemoveMember(db *mgo.Database, room models.Room, memberID string) *conf.ApiError {
  if err := db.C(c.mongo.RoomCollectionName).Update(
    bson.M{"_id": room.ID, "members": memberID},
    bson.M{
      "$pull": bson.M{"members": memberID},
      "$set":  bson.M{"updated": time.Now()},
    }); err != nil {
    return parseMongoDBError(err)
  }
  if len(room.Members) <= 2 {
    if err := db.C(c.mongo.RoomCollectionName).RemoveId(room.ID); err != nil {
      return parseMongoDBError(err)
    }
  }
  return nil
}

// Adds message to chat log to a given chat room by a given member.
// Additionally checks that current user is a member of that chat room.
func (c *MongoChat) AddMessage(message models.Message) (*models.Message, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if !bson.IsObjectIdHex(message.Room) {
    return nil, conf.ErrInvalidId
  }

  if count, err := db.C(c.mongo.RoomCollectionName).Find(bson.M{"_id": message.Room, "members": message.From}).Count();
    err != nil || count == 0 {
    return nil, conf.ErrNotAMember
  }

  log.Println("Log message from", message.From, "to room", message.Room)

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

  if count, err := db.C(c.mongo.RoomCollectionName).Find(bson.M{"_id": roomID, "members": profileID}).Count();
    err != nil || count == 0 {
    return nil, conf.ErrNotAMember
  }

  log.Println("Retrieving chat log from:", from, "limit:", limit, "room:", roomID, "member:", profileID)

  var messages []models.Message
  if err := db.C(c.mongo.MessagesCollectionName).Find(
    bson.M{"room": roomID, "timestamp": bson.M{"$lt": from}}).Limit(limit).Sort("-timestamp").All(&messages); err != nil {
    return nil, parseMongoDBError(err)
  }
  return messages, nil
}

func (c *MongoChat) AddReaction(profileID string, roomID string, messageID string, reaction string) *conf.ApiError {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.mongo.DBName)

  if count, err := db.C(c.mongo.RoomCollectionName).Find(bson.M{"_id": roomID, "members": profileID}).Count();
    err != nil || count == 0 {
    return conf.ErrNotAMember
  }

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

func parseMongoDBError(err error) *conf.ApiError {
  if err.Error() == "not found" {
    return conf.ErrNotFound
  } else if strings.Contains(err.Error(), "duplicate key error") {
    return conf.ErrAlreadyExists
  } else {
    return conf.NewApiError(err)
  }
}
