package services

import (
  "conf"
  "models"
)

type Account interface {
  GetProfile(accessToken string) (*models.User, *conf.ApiError)
  GetUser(profileID string) (*models.User, *conf.ApiError)
}
