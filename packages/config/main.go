package config

import (
	"encoding/json"
	"errors"

	"git.mills.io/prologic/bitcask"
)

func C() Data {
	db, err := bitcask.Open("/tmp/db/genocide")
	if errors.Is(err, bitcask.ErrDatabaseLocked) {
		panic(err)
	}

	var DB = Database{DB: db}

	//

	_, err = DB.DB.Get([]byte("accounts"))
	if errors.Is(err, bitcask.ErrKeyNotFound) {
		DB.DB.Put([]byte("accounts"), Struct2Bytes([]string{}))
	}

	_, err = DB.DB.Get([]byte("proxys"))
	if errors.Is(err, bitcask.ErrKeyNotFound) {
		DB.DB.Put([]byte("proxys"), Struct2Bytes([]Proxys{}))
	}

	_, err = DB.DB.Get([]byte("config"))
	if errors.Is(err, bitcask.ErrKeyNotFound) {
		DB.DB.Put([]byte("config"), Struct2Bytes(Config{}))
	}

	//

	var C Config
	var A []string
	var P []Proxys

	c, err := DB.DB.Get([]byte("config"))
	if err == nil {
		json.Unmarshal(c, &C)
	}
	a, err := DB.DB.Get([]byte("accounts"))
	if err == nil {
		json.Unmarshal(a, &A)
	}
	p, err := DB.DB.Get([]byte("proxys"))
	if err == nil {
		json.Unmarshal(p, &P)
	}

	return Data{
		Config:   C,
		Accounts: A,
		Proxys:   P,
		DB:       DB,
	}
}

func Struct2Bytes(L any) []byte {
	Body, _ := json.Marshal(L)
	return Body
}
