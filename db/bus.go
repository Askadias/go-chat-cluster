package db

type Bus interface {

  Receive() chan map[string][]byte

  Subscribe(id string) error

  Unsubscribe(id string) error

  Publish(id string, msg []byte) error
}
