package db

import (
  "github.com/Askadias/go-chat-cluster/models"
  "github.com/Askadias/go-chat-cluster/conf"
)

type Chat interface {
  OpenedRoomsCount(memberID string) (int, *conf.ApiError)
  CreateRoom(room models.Room) (*models.Room, *conf.ApiError)
  GetRooms(memberID string) ([]models.Room, *conf.ApiError)
  GetRoomsIn(roomIDs []string) ([]models.Room, *conf.ApiError)
  GetRoom(memberID string, roomID string) (*models.Room, *conf.ApiError)
  DeleteRoom(roomID string) *conf.ApiError
  AddRoomMember(roomID string, memberID string) *conf.ApiError
  RemoveRoomMember(roomID string, memberID string) *conf.ApiError
  IsRoomMember(roomID string, memberID string) *conf.ApiError
}
