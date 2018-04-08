package db

import "github.com/Askadias/go-chat-cluster/models"

type RoomCache interface {
  PutRoom(key string, room *models.Room) error
  GetRoom(key string) (*models.Room, error)
  EvictRoom(key string) error
}
