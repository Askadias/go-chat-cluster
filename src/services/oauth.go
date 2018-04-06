package services

import "github.com/Askadias/go-chat-cluster/src/conf"

type OAuth interface {
  ExchangeCodeToToken(accessCode string) (string, *conf.ApiError)
}
