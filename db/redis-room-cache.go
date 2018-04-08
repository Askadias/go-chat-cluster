package db

import (
  "github.com/garyburd/redigo/redis"
  "encoding/json"
  "github.com/Askadias/go-chat-cluster/conf"
  "log"
  "time"
  "github.com/Askadias/go-chat-cluster/models"
)

type RedisCacheOptions struct {
  RedisPool *redis.Pool
  CacheTTL  time.Duration
}

type RedisCache struct {
  options RedisCacheOptions
}

func NewRedisCache(options RedisCacheOptions) *RedisCache {
  cache := &RedisCache{
    options: options,
  }
  return cache
}

func (c *RedisCache) PutRoom(key string, room *models.Room) error {
  conn := c.options.RedisPool.Get()
  defer conn.Close()
  if data, err := json.Marshal(room); err != nil {
    log.Println("Failed to marshal room for redis cache:", err, "Key:", key)
    c.EvictRoom(key)
    return conf.NewApiError(err)
  } else {
    if _, err := conn.Do("SET", key, data, "EX", c.options.CacheTTL.Seconds()); err != nil {
      log.Println("Failed to put room to redis cache:", err, "Key:", key)
      c.EvictRoom(key)
      return conf.NewApiError(err)
    }
    return nil
  }
}

func (c *RedisCache) GetRoom(key string) (*models.Room, error) {
  conn := c.options.RedisPool.Get()
  defer conn.Close()
  if data, err := conn.Do("GET", key); err != nil {
    log.Println("Failed to get room from redis cache:", err, "Key:", key)
    return nil, conf.NewApiError(err)
  } else {
    if data == nil {
      return nil, conf.ErrNotFound
    }
    var room models.Room
    if err := json.Unmarshal(data.([]byte), &room); err != nil {
      log.Println("Failed to unmarshal room for redis cache:", err, "Key:", key)
      return nil, conf.NewApiError(err)
    } else {
      return &room, nil
    }
  }
}

func (c *RedisCache) EvictRoom(key string) error {
  conn := c.options.RedisPool.Get()
  defer conn.Close()
  if _, err := conn.Do("DEL", key); err != nil {
    log.Println("Failed to delete room from redis cache:", err, "Key:", key)
    return conf.NewApiError(err)
  } else {
    return nil
  }
}
