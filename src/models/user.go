package models

type User struct {
  Id        string `json:"id" bson:"id"`
  Name      string `json:"name" bson:"name"`
  AvatarURL string `json:"avatarUrl,omitempty" bson:"-"`
}

type UserList struct {
  Data []User `json:"data"`
}
