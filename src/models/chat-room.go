package models

import (
  "gopkg.in/mgo.v2/bson"
  "time"
)

type ChatRoom struct {
  Id        bson.ObjectId `json:"id" bson:"_id"`
  OwnerId   string        `json:"ownerId" bson:"ownerId"`
  Members   []User        `json:"members" bson:"members"`
  Created   time.Time     `json:"created" bson:"created"`
  Updated   time.Time     `json:"updated" bson:"updated"`
}
