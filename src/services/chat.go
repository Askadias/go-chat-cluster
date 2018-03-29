package services

import (
  "gopkg.in/mgo.v2"
  "crypto/tls"
  "net"
  "log"
  "time"
  "models"
  "conf"
  "gopkg.in/mgo.v2/bson"
  "strings"
)

type ChatDBOptions struct {
  MongoURL     string
  MongoDBName  string
  MongoTimeout time.Duration
  Facebook     *Facebook
}

// Chat Service manages chat rooms and their members.
type Chat struct {
  mongoSession *mgo.Session
  options      ChatDBOptions
}

// Creates Chat Service. Initialises MongoDB connection.
func NewChat(options ChatDBOptions) *Chat {
  dialInfo, err := mgo.ParseURL(options.MongoURL)

  tlsConfig := &tls.Config{}
  dialInfo.Timeout = options.MongoTimeout
  dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
    conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
    return conn, err
  }
  session, err := mgo.DialWithInfo(dialInfo)
  if err != nil {
    log.Fatalf("CreateSession: %s\n", err)
  }
  return &Chat{
    mongoSession: session,
    options:      options,
  }
}

// Save new ChatRoom into the MongoDB by setting current profile ID as an owner.
// Automatically adds current user to the chat room members.
func (c *Chat) CreateRoom(profileID string, room models.ChatRoom) (*models.ChatRoom, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.options.MongoDBName)

  room.OwnerId = profileID
  room.Id = bson.NewObjectId()
  now := time.Now()
  room.Created = now
  room.Updated = now

  if err := db.C("chat-room").Insert(room); err != nil {
    return nil, parseMongoDBError(err)
  }
  return &room, nil
}

// Returns all the ChatRooms by a given member
func (c *Chat) GetRooms(profileID string) ([]models.ChatRoom, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.options.MongoDBName)

  log.Println("Retrieving chat rooms for", profileID)

  var rooms []models.ChatRoom
  if err := db.C("chat-room").Find(bson.M{"members.id": profileID}).All(&rooms); err != nil {
    return nil, conf.NewApiError(err)
  }
  return rooms, nil
}

// Returns ChatRoom by ID
func (c *Chat) GetRoom(profileID string, roomId bson.ObjectId) (*models.ChatRoom, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.options.MongoDBName)

  log.Println("Retrieving chat room", roomId, "by", profileID)

  var room models.ChatRoom
  if err := db.C("chat-room").FindId(roomId).One(&room); err != nil {
    return nil, parseMongoDBError(err)
  }
  return &room, nil
}

// Deletes ChatRoom from the MongoDB by its owner.
func (c *Chat) DeleteRoom(profileID string, roomId bson.ObjectId) *conf.ApiError {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.options.MongoDBName)

  var room models.ChatRoom
  if err := db.C("chat-room").FindId(roomId).One(&room); err != nil {
    return parseMongoDBError(err)
  }
  if room.OwnerId != profileID {
    return conf.ErrNotAnOwner
  }

  log.Println("Deleting chat room", roomId, "by owner", profileID)
  if err := db.C("chat-room").RemoveId(roomId); err != nil {
    return parseMongoDBError(err)
  }
  return nil
}

// Updates current ChatRoom by adding new member.
// Actual member info retrieved from facebook.
// Additionally checks that current user is an owner of that room to update it.
func (c *Chat) AddRoomMember(profileID string, roomId bson.ObjectId, memberId string) *conf.ApiError {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.options.MongoDBName)

  var room models.ChatRoom
  if err := db.C("chat-room").FindId(roomId).One(&room); err != nil {
    return parseMongoDBError(err)
  }
  if room.OwnerId != profileID {
    return conf.ErrNotAnOwner
  }

  actualUser, err := c.options.Facebook.GetUser(memberId)
  if err != nil {
    return conf.ErrNoProfile
  }

  log.Println("Add member", memberId, "to room", roomId, "by", profileID)

  if err := db.C("chat-room").Update(
    bson.M{"_id": roomId, "members.id": bson.M{"$ne": memberId}},
    bson.M{
      "$push": bson.M{"members": *actualUser},
      "$set":  bson.M{"updated": time.Now()},
    }); err != nil {
    return parseMongoDBError(err)
  }
  return nil
}

// Updates current ChatRoom by removing a member with a specific id.
// Additionally checks that current user is an owner of that room to update it.
func (c *Chat) RemoveRoomMember(profileID string, roomId bson.ObjectId, memberId string) *conf.ApiError {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.options.MongoDBName)

  var room models.ChatRoom
  if err := db.C("chat-room").FindId(roomId).One(&room); err != nil {
    return parseMongoDBError(err)
  }
  if room.OwnerId != profileID {
    return conf.ErrNotAnOwner
  }

  log.Println("Remove member", memberId, "from room", roomId, "by", profileID)

  if err := db.C("chat-room").UpdateId(
    roomId,
    bson.M{
      "$pull": bson.M{"members": bson.M{"id": memberId}},
      "$set":  bson.M{"updated": time.Now()},
    }); err != nil {
    return parseMongoDBError(err)
  }
  return nil
}

// Adds message to chat log to a given chat room by a given member.
// Additionally checks that current user is a member of that chat room.
func (c *Chat) LogMessage(profileID string, roomId bson.ObjectId, message models.Message) (*models.Message, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.options.MongoDBName)

  log.Println("Log message from", profileID, "to roomId", roomId)
  if count, err := db.C("chat-room").Find(bson.M{"members.id": profileID, "_id": roomId}).Count();
    err != nil || count == 0 {
    return nil, conf.ErrNotAMember
  }

  message.Id = bson.NewObjectId()
  message.RoomId = roomId
  message.From = profileID
  message.Timestamp = time.Now()

  if err := db.C("chat-log").Insert(message); err != nil {
    return nil, parseMongoDBError(err)
  }
  return &message, nil
}

// Returns chat log page by a given start timestamp and limit.
// Additionally checks that current user is a member of that chat room.
func (c *Chat) GetChatLog(profileID string, roomId bson.ObjectId, from time.Time, limit int) ([]models.Message, *conf.ApiError) {
  s := c.mongoSession.Clone()
  defer s.Close()
  db := s.DB(c.options.MongoDBName)

  log.Println("Retrieving chat log from:", from, "limit:", limit, "roomId:", roomId, "member:", profileID)

  count, err := db.C("chat-room").Find(bson.M{"members.id": profileID, "_id": roomId}).Count()
  if err != nil || count == 0 {
    return nil, conf.ErrNotAMember
  }

  var messages []models.Message
  if err := db.C("chat-log").Find(
    bson.M{"roomId": roomId, "timestamp": bson.M{"$lt": from}}).Limit(limit).All(&messages); err != nil {
    return nil, parseMongoDBError(err)
  }
  return messages, nil
}

func parseMongoDBError(err error) *conf.ApiError {
  if err.Error() == "not found" {
    return conf.ErrNotFound
  } else if strings.Contains(err.Error(), "duplicate key error"){
    return conf.ErrAlreadyExists
  } else {
    return conf.NewApiError(err)
  }
}
