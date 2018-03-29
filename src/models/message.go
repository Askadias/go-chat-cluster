package models

import (
  "time"
  "gopkg.in/mgo.v2/bson"
)

type Message struct {
  Id        bson.ObjectId `json:"-" bson:"_id"`
  RoomId    bson.ObjectId `json:"roomId" bson:"roomId"`
  From      string        `json:"from" bson:"from"`
  Timestamp time.Time     `json:"timestamp" bson:"timestamp"`
  Body      string        `json:"body" bson:"body"`
}
