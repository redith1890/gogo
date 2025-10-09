package main

import (
	"bytes"
	"encoding/gob"
	"go.etcd.io/bbolt"
	"time"
	."fmt"
	"encoding/binary"
)

type Player struct {
	Id          uint64
	Name        string
	InGame      bool
	LastConnect time.Time
}

type Game struct {
	Id uint64
	PlayerId1 uint64
	PlayerId2 uint64
}

type Transaction = *bbolt.Tx

func itob(id uint64) []byte {
	var key [8]byte
	binary.BigEndian.PutUint64(key[:], id)
	return key[:]
}

func CreatePlayer(p Player) (uint64, error) {
	var id uint64
	return id, DB.Update(func(tx Transaction) error {
		b := tx.Bucket([]byte("player"))

		id ,_ = b.NextSequence()
		p.Id = id

		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(p); err != nil {
			return err
		}

		return b.Put(itob(id), buf.Bytes())
	})
}

func GetPlayer(name string) (*Player, error) {
	var p Player
	err := DB.View(func(tx Transaction) error {
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

func CreateGame(game Game) (uint64, error) {
	var id uint64
	return id, DB.Update(func(tx Transaction) error {
		b := tx.Bucket([]byte("game"))
		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(game); err != nil {
			return err
		}
		id, _ = b.NextSequence()

		return b.Put(itob(game.Id), buf.Bytes())
	})
}

func GetLastGameId() (id uint64) {
	var game Game
	DB.View(func(tx Transaction) error {
		b := tx.Bucket([]byte("game"))
		c := b.Cursor()
		k, v := c.Last()
		Println(k)
		Println(v)
		return nil
	})


	return game.Id
}