package db

import (
  "github.com/Askadias/go-chat-cluster/models"
  "github.com/Askadias/go-chat-cluster/conf"
  "time"
)

type ChatLog interface {
  AddMessage(message models.Message) (*models.Message, *conf.ApiError)
  GetMessages(profileID string, roomID string, from time.Time, limit int) ([]models.Message, *conf.ApiError)
  AddReaction(profileId string, roomID string, messageID string, reaction string) *conf.ApiError
  EditMessage(profileId string, messageID string, body string) *conf.ApiError
}
