package db

import (
  "models"
  "conf"
)

type Chat interface {
  CreateRoom(profileID string, room models.Room) (*models.Room, *conf.ApiError)
  GetRooms(profileID string) ([]models.Room, *conf.ApiError)
  GetRoom(profileID string, roomID string) (*models.Room, *conf.ApiError)
  DeleteRoom(profileID string, roomID string) *conf.ApiError
  AddRoomMember(profileID string, roomID string, memberID string) *conf.ApiError
  RemoveRoomMember(profileID string, roomID string, memberID string) *conf.ApiError
}
