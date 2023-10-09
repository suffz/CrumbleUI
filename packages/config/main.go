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
		DB.DB.Put([]byte("config"), Struct2Bytes(Config{
			Settings: AccountSettings{
				GC_ReqAmt:       1,
				MFA_ReqAmt:      1,
				AccountsPerGc:   5,
				AccountsPerMfa:  1,
				SleepAmtPerGc:   15000,
				SleepAmtPerMfa:  10000,
				Spread:          0,
				UseCustomSpread: false,
			},
			Bools: Bools{
				FirstUse:           true,
				UseProxyDuringAuth: false,
				DownloadedPW:       false,
			},
			SkinChange: Skin{
				Variant: "slim",
				Link:    "https://textures.minecraft.net/texture/516accb84322ca168a8cd06b4d8cc28e08b31cb0555eee01b64f9175cefe7b75",
			},
		}))
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
	l_o_l, err := DB.DB.Get([]byte("discord_id"))
	if err == nil {
		json.Unmarshal(p, &P)
	}
	b_o_o_m_e_r, err := DB.DB.Get([]byte("discord_avatar"))
	if err == nil {
		json.Unmarshal(p, &P)
	}

	return Data{
		Config:            C,
		Accounts:          A,
		Proxys:            P,
		DB:                DB,
		DiscordID:         string(l_o_l),
		DiscordImageBytes: string(b_o_o_m_e_r),
	}
}

func Struct2Bytes(L any) []byte {
	Body, _ := json.Marshal(L)
	return Body
}
