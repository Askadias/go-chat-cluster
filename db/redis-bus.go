package db

import (
  "github.com/garyburd/redigo/redis"
  "log"
  "github.com/Askadias/go-chat-cluster/conf"
)

type RedisOptions struct {
  RedisPool *redis.Pool
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
  if conn := bus.options.RedisPool.Get(); conn.Err() != nil {
    panic(conn.Err())
  } else {
    bus.pubSubConn = &redis.PubSubConn{Conn: conn}
    defer bus.pubSubConn.Close()
    defer conn.Close()
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
}

func (bus *RedisBus) Receive() chan map[string][]byte {
  return bus.receive
}

func (bus *RedisBus) Subscribe(id string) error {
  if bus.pubSubConn != nil {
    return bus.pubSubConn.Subscribe(id)
  } else {
    return conf.ErrNotInitialized
  }
}

func (bus *RedisBus) Unsubscribe(id string) error {
  return bus.pubSubConn.Unsubscribe(id)
}

func (bus *RedisBus) Publish(id string, msg []byte) error {
  if conn := bus.options.RedisPool.Get(); conn.Err() != nil {
    log.Printf("error on getting redis connection from pool. %s\n", conn.Err())
    return conn.Err()
  } else {
    defer conn.Close()
    conn.Do("PUBLISH", id, msg)
    return nil
  }
}
