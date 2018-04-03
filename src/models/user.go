package models

type User struct {
  ID        string `json:"id"`
  Name      string `json:"name"`
  AvatarURL string `json:"avatarUrl,omitempty"`
}

type UserList struct {
  Data []User `json:"data"`
}
