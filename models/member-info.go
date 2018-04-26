package models

// Structure describes particular room membership
// I.e. if user exits private chat the second participant still can see the chat log and current member
// while current user don't see the chat until its participant starts writing into it by initiating
// new conversation fur the current user
type MemberInfo struct {
  ID         string    `json:"id" bson:"_id"`
  RoomID     string    `json:"room" bson:"room"`
  MemberID   string    `json:"member" bson:"member"`
  JoinedAt   Timestamp `json:"joinedAt" bson:"joinedAt"`     // marker to start reading messages from
  LastReadAt Timestamp `json:"lastReadAt" bson:"lastReadAt"` // marker to notify about the new messages
}
