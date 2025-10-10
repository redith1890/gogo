package main

import (
	// . "fmt"
	bolt "go.etcd.io/bbolt"
	"log"
	"sync"
	"time"
	"golang.org/x/net/websocket"
)

var DB *bolt.DB

func InitDB() error {
	path := "my.db"
	var err error
	DB, err = bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	err = DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("user"))
		_, err = tx.CreateBucketIfNotExists([]byte("game"))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
	return err
}

type Session struct {
	ID        string
	Values    map[string]string
	ExpiresAt time.Time
}

type SessionStore struct {
	Sessions map[string]Session
	Mu       sync.RWMutex
}

type Server struct {
	conns   map[*websocket.Conn]*SessionInfo
	Changed chan struct{}
}

var Store = &SessionStore{
	Sessions: make(map[string]Session),
}

//Temporal if we want to do more servers
var MainServer *Server

type SessionInfo struct {
	Id  uint64
	Flags ConnFlag
}
