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

func btoi(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func DecodeGob(data []byte, out interface{}) error {
	if data == nil {
		return nil
	}
	return gob.NewDecoder(bytes.NewReader(data)).Decode(out)
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

func GetPlayerById(id uint64) *Player {
	var p Player
	DB.View(func(tx Transaction) error {
		b := tx.Bucket([]byte("player"))
		v := b.Get([]byte(itob(id)))
		if v == nil {
			return nil
		}
		return DecodeGob(v, &p)
	})
	return &p
}

func GetPlayerByUsername(username string) *Player {
	var p Player
	found := false
	DB.View(func(tx Transaction) error {
		b := tx.Bucket([]byte("player"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			DecodeGob(v, &p)
			if p.Name == username {
				found = true
				break
			}
		}
		return nil
	})

	if !found {
		return nil
	}
	return &p
}

func GetAllPlayersIds() []uint64 {
	var ids []uint64

	DB.View(func(tx Transaction) error {
		b := tx.Bucket([]byte("player"))
		// TODO batch this
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			ids = append(ids, btoi(k))
		}
		return nil
	})
	return ids
}

func GetAllPlayersUsernames() []string {
	var usernames []string

	DB.View(func(tx Transaction) error {
		b := tx.Bucket([]byte("player"))
		// TODO batch this
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var player Player
			DecodeGob(v, &player)
			usernames = append(usernames, player.Name)
		}
		return nil
	})
	return usernames
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