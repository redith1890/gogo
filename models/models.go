package models

import (
	"go.etcd.io/bbolt"
	"time"
	. "go-online/globals"
	"bytes"
	"encoding/gob"
)

type Player struct {
	Name        string
	InGame      bool
	LastConnect time.Time
}

func SavePlayer(p Player) error {
	return DB.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("player"))
		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(p); err != nil {
			return err
		}
		return b.Put([]byte(p.Name), buf.Bytes())
	})
}

func GetPlayer(name string) (*Player, error) {
	var p Player
	err := DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("player"))
		v := b.Get([]byte(name))
		if v == nil {
			return nil
		}
		return gob.NewDecoder(bytes.NewReader(v)).Decode(&p)
	})
	if err != nil {
		return nil, err
	}
	if p.Name == "" {
		return nil, nil
	}
	return &p, nil
}
