package db

import (
  "github.com/Askadias/go-chat-cluster/models"
  "github.com/Askadias/go-chat-cluster/conf"
)

type Chat interface {
  CreateRoom(profileID string, room models.Room) (*models.Room, *conf.ApiError)
  GetRooms(profileID string) ([]models.Room, *conf.ApiError)
  GetRoom(profileID string, roomID string) (*models.Room, *conf.ApiError)
  DeleteRoom(profileID string, roomID string) (*models.Room, *conf.ApiError)
  AddRoomMember(profileID string, roomID string, memberID string) (*models.Room, *conf.ApiError)
  RemoveRoomMember(profileID string, roomID string, memberID string) (*models.Room, *conf.ApiError)
}
