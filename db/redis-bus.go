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
  bus := &RedisBus{
    options: options,
    receive: make(chan map[string][]byte),
  }
  go bus.run()
  return bus
}

func (bus *RedisBus) run() {
  redisConn, err := bus.conn()
  if err != nil {
    panic(err)
  }
  defer redisConn.Close()
  bus.pubSubConn = &redis.PubSubConn{Conn: redisConn}
  defer bus.pubSubConn.Close()
  for {
    switch v := bus.pubSubConn.Receive().(type) {
    case redis.Message:
      msg := make(map[string][]byte)
      msg[v.Channel] = v.Data
      bus.receive <- msg

    case redis.Subscription:
      log.Printf("subscription message: %s: %s %d\n", v.Channel, v.Kind, v.Count)

    case error:
      log.Println("error pub/sub, delivery has stopped")
      return
    }
  }
}

func (bus *RedisBus) Receive() chan map[string][]byte {
  return bus.receive
}
func (bus *RedisBus) Subscribe(id string) error {
  return bus.pubSubConn.Subscribe(id)
}

func (bus *RedisBus) Unsubscribe(id string) error {
  return bus.pubSubConn.Unsubscribe(id)
}

func (bus *RedisBus) Publish(id string, msg []byte) error {
  if c, err := bus.conn(); err != nil {
    log.Printf("error on redis conn. %s\n", err)
    return err
  } else {
    c.Do("PUBLISH", id, msg)
    return nil
  }
}

func (bus *RedisBus) conn() (redis.Conn, error) {
  return redis.DialURL(bus.options.RedisURL)
}
