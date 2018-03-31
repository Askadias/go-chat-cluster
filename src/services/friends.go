package services

import (
  "models"
  "conf"
)

type Friends interface {
  GetFriends(profileID string) ([]models.User, *conf.ApiError)
}
