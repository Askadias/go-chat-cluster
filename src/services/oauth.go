package services

import "conf"

type OAuth interface {
  ExchangeCodeToToken(accessCode string) (string, *conf.ApiError)
}
