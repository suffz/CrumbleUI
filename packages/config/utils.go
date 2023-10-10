package config

import (
	"main/packages/apiGO"
	"main/packages/h2"
	"math/rand"
	"strings"
	"time"
)

func CheckDupes(strs []apiGO.Info) []apiGO.Info {
	dedup := strs[:0]
	track := make(map[string]bool, len(strs))

	for _, str := range strs {
		if track[str.Email+":"+str.Password] {
			continue
		}
		dedup = append(dedup, str)
		track[str.Email+":"+str.Password] = true
	}

	return dedup
}

func RandProxyUse(proxys []string) (Use *h2.ProxyAuth) {

	if len(proxys) == 0 {
		return nil
	}

	n := rand.New(rand.NewSource(time.Now().UnixMicro()))
	if data := strings.Split(proxys[n.Intn(len(proxys))], ":"); len(data) > 0 {
		ip, port, user, password := data[0], data[1], "", ""
		if len(data) > 2 {
			user, password = data[2], data[3]
		}
		Use = &h2.ProxyAuth{IP: ip, Port: port, User: user, Password: password}
	}
	return
}

func (DB *Data) SaveProxys(Temp []Proxys) {
	if err := DB.DB.DB.Put([]byte(`proxys`), Struct2Bytes(Temp)); err != nil {
		panic(err)
	}
}

func (DB *Data) SaveConfig() {
	if err := DB.DB.DB.Put([]byte(`config`), Struct2Bytes(DB.Config)); err != nil {
		panic(err)
	}
}

/*

func (DB *Data) GetAllAccountValues() (ACCS []Logs) {
	T, _ := DB.DB.DB.Get([]byte(`accounts`))
	json.Unmarshal(T, &ACCS)
	fmt.Println(ACCS)
	return
}

*/

/*

func (DB *Data) DeleteAccount(Name string) {
	var ACCS []Logs
	T, _ := DB.DB.DB.Get([]byte(`accounts`))
	json.Unmarshal(T, &ACCS)
	var New []Logs
	for _, a := range ACCS {
		if !strings.EqualFold(a.Name, Name) {
			New = append(New, a)
		}
	}
	DB.DB.DB.Put([]byte(`accounts`), Struct2Bytes(New))
}

*/

func ProxyData(p string) Proxys {
	data := strings.Split(p, ":")
	switch len(data) {
	case 4:
		return Proxys{
			Ip:       data[0],
			Port:     data[1],
			User:     data[2],
			Password: data[3],
		}
	case 2:
		return Proxys{
			Ip:   data[0],
			Port: data[1],
		}
	}
	return Proxys{}
}

type Logs struct {
	Email, Password string
	Name            string
	Bearer          string
	Info            Info
}

type Info struct {
	ID   string
	Name string
}
