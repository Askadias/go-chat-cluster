package db

import (
  "models"
  "conf"
  "time"
)

type ChatLog interface {
  AddMessage(message models.Message) (*models.Message, *conf.ApiError)
  GetMessages(profileID string, roomID string, from time.Time, limit int) ([]models.Message, *conf.ApiError)
}
