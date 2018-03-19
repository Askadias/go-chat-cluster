package models

type User struct {
  Id        string `json:"id"`
  Name      string `json:"name"`
  AvatarURL string `json:"avatarUrl"`
}

type UserList struct {
  Data []User `json:"data"`
}
