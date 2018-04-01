package db

import (
  "github.com/garyburd/redigo/redis"
  "log"
)

type RedisOptions struct {
  RedisURL string
}

type RedisBus struct {
  options    RedisOptions
  pubSubConn *redis.PubSubConn
  receive    chan map[string][]byte
}

func NewRedisBus(options RedisOptions) *RedisBus {
  b := &RedisBus{
    options: options,
    receive: make(chan map[string][]byte),
  }
  go b.run()
  return b
}

func (b *RedisBus) run() {
  redisConn, err := b.conn()
  if err != nil {
    panic(err)
  }
  defer redisConn.Close()
  b.pubSubConn = &redis.PubSubConn{Conn: redisConn}
  defer b.pubSubConn.Close()
  for {
    switch v := b.pubSubConn.Receive().(type) {
    case redis.Message:
      msg := make(map[string][]byte)
      msg[v.Channel] = v.Data
      b.receive <- msg

    case redis.Subscription:
      log.Printf("subscription message: %s: %s %d\n", v.Channel, v.Kind, v.Count)

    case error:
      log.Println("error pub/sub, delivery has stopped")
      return
    }
  }
}

func (b *RedisBus) Receive() chan map[string][]byte {
  return b.receive
}
func (b *RedisBus) Subscribe(id string) error {
  return b.pubSubConn.Subscribe(id)
}

func (b *RedisBus) Unsubscribe(id string) error {
  return b.pubSubConn.Unsubscribe(id)
}

func (b *RedisBus) Publish(id string, msg []byte) error {
  if c, err := b.conn(); err != nil {
    log.Printf("error on redis conn. %s\n", err)
    return err
  } else {
    c.Do("PUBLISH", id, msg)
    return nil
  }
}

func (b *RedisBus) conn() (redis.Conn, error) {
  return redis.DialURL(b.options.RedisURL)
}
