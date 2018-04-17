package models

type Message struct {
  ID        string            `json:"id" bson:"_id"`
  RoomID    string            `json:"room" bson:"room"`
  FromID    string            `json:"from" bson:"from"`
  Timestamp Timestamp         `json:"timestamp" bson:"timestamp"`
  Type      string            `json:"type" bson:"-"`
  Body      string            `json:"body" bson:"body"`
  Reactions map[string]string `json:"reactions,omitempty" bson:"reactions,omitempty"`
}
