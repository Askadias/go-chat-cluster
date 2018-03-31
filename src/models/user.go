package models

type User struct {
  ID        string `json:"id" bson:"id"`
  Name      string `json:"name" bson:"name"`
  AvatarURL string `json:"avatarUrl,omitempty" bson:"avatarUrl,omitempty"`
}

type UserList struct {
  Data []User `json:"data"`
}
