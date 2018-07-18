package db

import (
  "github.com/garyburd/redigo/redis"
  "time"
  "github.com/Askadias/go-chat-cluster/conf"
)

func NewRedisPool(options conf.RedisConf) *redis.Pool {

  return &redis.Pool{
    MaxIdle:     options.MaxIdle,
    MaxActive:   options.MaxActive,
    IdleTimeout: options.IdleTimeout,
    TestOnBorrow: func(c redis.Conn, t time.Time) error {
      if time.Since(t) < time.Minute {
        return nil
      }
      _, err := c.Do("PING")
      return err
    },
    Dial: func() (redis.Conn, error) { return redis.DialURL(options.URL) },
  }
}
