package services

import (
  "github.com/Askadias/go-chat-cluster/models"
  "github.com/Askadias/go-chat-cluster/conf"
)

type Friends interface {
  GetFriends(profileID string) ([]models.User, *conf.ApiError)
}
