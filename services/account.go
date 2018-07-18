package services

import (
  "github.com/Askadias/go-chat-cluster/conf"
  "github.com/Askadias/go-chat-cluster/models"
)

type Account interface {

  GetProfile(accessToken string) (*models.User, *conf.ApiError)

  GetUser(profileID string) (*models.User, *conf.ApiError)
}
