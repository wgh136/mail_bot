package mail

import (
	"fmt"
	"log"
	"time"
)

type Status int

const (
	StatusDisconnected Status = iota
	StatusConnecting
	StatusConnected
	StatusError
)

type Listener func(uid int64, from string, to string, subject string, body string)

type Connection interface {
	Connect() error

	Disconnect()

	Status() Status

	GetUserID() int64

	GetEmail() string

	AddListener(listener Listener)
}

type ConnectionKey struct {
	UserID int64
	Email  string
}

var connections = make(map[ConnectionKey]Connection)

func init() {
	go func() {
		time.Sleep(5 * time.Minute)
		for _, c := range connections {
			if c.Status() == StatusDisconnected || c.Status() == StatusError {
				_ = c.Connect()
			}
		}
	}()
}

func AddConnection(c Connection) error {
	key := ConnectionKey{
		UserID: c.GetUserID(),
		Email:  c.GetEmail(),
	}
	_, exists := connections[key]
	if exists {
		return fmt.Errorf("connection already exists for user %d and email %s", c.GetUserID(), c.GetEmail())
	}
	err := c.Connect()
	if err != nil {
		return err
	}
	c.AddListener(func(uid int64, from string, to string, subject string, body string) {
		log.Println("New email received:", uid, from, to, subject)
	})
	connections[key] = c
	return nil
}

func RemoveConnection(uid int64, email string) {
	key := ConnectionKey{
		UserID: uid,
		Email:  email,
	}
	c, exists := connections[key]
	if !exists {
		log.Println("No connection found for user", uid, "and email", email)
		return
	}
	c.Disconnect()
	delete(connections, key)
	log.Println("Connection removed for user", uid, "and email", email)
}
