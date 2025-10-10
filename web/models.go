package main

import (
	"bytes"
	"encoding/gob"
	"go.etcd.io/bbolt"
	"time"
	."fmt"
	"encoding/binary"
)

type User struct {
	Id          uint64
	Name        string
	InGame      bool
	LastConnect time.Time
}

type Game struct {
	Id uint64
	UserId1 uint64
	UserId2 uint64
}

type Transaction = *bbolt.Tx

func (Game) Print() {
	Println()
}

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

func GetAllGames() []Game {
	var games []Game

	DB.View(func(tx Transaction) error {
		b := tx.Bucket([]byte("game"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var game Game
			DecodeGob(v, &game)
			games = append(games, game)
		}
		return nil
	})
	return games
}

func CreateUser(p User) (uint64, error) {
	var id uint64
	return id, DB.Update(func(tx Transaction) error {
		b := tx.Bucket([]byte("user"))

		id ,_ = b.NextSequence()
		p.Id = id

		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(p); err != nil {
			return err
		}

		return b.Put(itob(id), buf.Bytes())
	})
}

func GetUserById(id uint64) *User {
	var p User
	DB.View(func(tx Transaction) error {
		b := tx.Bucket([]byte("user"))
		v := b.Get([]byte(itob(id)))
		if v == nil {
			return nil
		}
		return DecodeGob(v, &p)
	})
	return &p
}

func GetUserByUsername(username string) *User {
	var p User
	found := false
	DB.View(func(tx Transaction) error {
		b := tx.Bucket([]byte("user"))
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

func GetAllUsersIds() []uint64 {
	var ids []uint64

	DB.View(func(tx Transaction) error {
		b := tx.Bucket([]byte("user"))
		// TODO batch this
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			ids = append(ids, btoi(k))
		}
		return nil
	})
	return ids
}

func GetAllUsersUsernames() []string {
	var usernames []string

	DB.View(func(tx Transaction) error {
		b := tx.Bucket([]byte("user"))
		// TODO batch this
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var user User
			DecodeGob(v, &user)
			usernames = append(usernames, user.Name)
		}
		return nil
	})
	return usernames
}

func CreateGame(game Game) (uint64, error) {
	var id uint64
	Printf("Game created id1 = %d, id2 = %d \n", game.UserId1, game.UserId2)
	err := DB.Update(func(tx Transaction) error {
		b := tx.Bucket([]byte("game"))
		buf := new(bytes.Buffer)
		id, _ = b.NextSequence()
		game.Id = id

		if err := gob.NewEncoder(buf).Encode(game); err != nil {
			return err
		}

		return b.Put(itob(id), buf.Bytes())
	})
	return id, err
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