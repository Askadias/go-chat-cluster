package services

import "github.com/Askadias/go-chat-cluster/conf"

type OAuth interface {
  ExchangeCodeToToken(accessCode string) (string, *conf.ApiError)
}
