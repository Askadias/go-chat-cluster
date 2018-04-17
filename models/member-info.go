package models

type MemberInfo struct {
  ID         string    `json:"id" bson:"_id"`
  RoomID     string    `json:"room" bson:"room"`
  MemberID   string    `json:"member" bson:"member"`
  JoinedAt   Timestamp `json:"joinedAt" bson:"joinedAt"`
  LastReadAt Timestamp `json:"lastReadAt" bson:"lastReadAt"`
}
