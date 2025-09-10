package globals

import (
	"time"
	"sync"
	bolt "go.etcd.io/bbolt"
	. "fmt"
	"log"
)

var DB *bolt.DB

func InitDB(){
	path := "my.db"
	var err error
	DB, err = bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		Println(err)
		return
	}
	err = DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("player"))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
}

type Session struct {
	ID        string
	Values    map[string]string
	ExpiresAt time.Time
}

type SessionStore struct {
	Sessions map[string]Session
	Mu sync.RWMutex
}

var Store = &SessionStore{
	Sessions: make(map[string]Session),
}
