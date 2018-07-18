package db

import (
  "github.com/Askadias/go-chat-cluster/models"
  "github.com/Askadias/go-chat-cluster/conf"
)

type MemberInfo interface {

  CreateMemberInfo(memberInfo models.MemberInfo) (*models.MemberInfo, *conf.ApiError)

  GetMemberInfo(roomID string, memberID string) (*models.MemberInfo, *conf.ApiError)

  GetAllMembersInfo(roomID string) ([]models.MemberInfo, *conf.ApiError)

  UpdateLastReadTime(roomID string, memberID string) *conf.ApiError

  DeleteMemberInfo(roomID string, memberID string) *conf.ApiError

  DeleteAllMembersInfo(roomID string) *conf.ApiError

  GetAllRoomsInfo(memberID string) ([]models.MemberInfo, *conf.ApiError)
}
