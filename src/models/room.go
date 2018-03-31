package models

type Room struct {
  ID      string    `json:"id" bson:"_id"`
  Owner   string    `json:"owner" bson:"owner"`
  Members []string  `json:"members" bson:"members"`
  Created Timestamp `json:"created" bson:"created"`
  Updated Timestamp `json:"updated" bson:"updated"`
}
