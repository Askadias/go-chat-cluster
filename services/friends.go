package services

import (
  "github.com/Askadias/go-chat-cluster/src/models"
  "github.com/Askadias/go-chat-cluster/src/conf"
)

type Friends interface {
  GetFriends(profileID string) ([]models.User, *conf.ApiError)
}
