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

func (DB *Data) SaveAccount(Temp []apiGO.Info, proxys []string) {

	var T []apiGO.Info

	for _, a := range DB.Config.Bearers {
		for _, Temp := range Temp {
			if (a.Email == Temp.Email && a.Password == Temp.Password) || (a.Info.ID == Temp.Info.ID) {
				T = append(T, Temp)
				break
			}
		}
	}

	if len(T) > 0 {
		Temp = CheckDupes(T)
	}

	for _, Temp := range Temp {

		var Found []int
		for i, a := range DB.Config.Bearers {
			if a.Info.Name == Temp.Info.Name {
				Found = append(Found, 1)
				DB.Config.Bearers[i].Bearer = Temp.Bearer
				DB.Config.Bearers[i].Email = Temp.Email
				DB.Config.Bearers[i].Password = Temp.Password
				DB.Config.Bearers[i].Info = UserINFO{
					Name: Temp.Info.Name,
					ID:   Temp.Info.ID,
				}
			}
		}
		if len(Found) == 0 {
			DB.Config.Bearers = append(DB.Config.Bearers, Bearers{
				Bearer:       Temp.Bearer,
				Email:        Temp.Email,
				Password:     Temp.Password,
				AuthInterval: 54000,
				AuthedAt:     time.Now().Unix(),
				Type:         Temp.AccountType,
				NameChange:   apiGO.CheckChange(Temp.Password, randProxyUse(proxys)),
				Info: UserINFO{
					Name: Temp.Info.Name,
					ID:   Temp.Info.ID,
				},
			})
		}
	}

	var New []Bearers
	for _, Temp := range Temp {
		New = append(New, Bearers{
			Bearer:       Temp.Bearer,
			Email:        Temp.Email,
			Password:     Temp.Password,
			AuthInterval: 54000,
			AuthedAt:     time.Now().Unix(),
			Type:         Temp.AccountType,
			NameChange:   apiGO.CheckChange(Temp.Bearer, randProxyUse(proxys)),
			Info: UserINFO{
				Name: Temp.Info.Name,
				ID:   Temp.Info.ID,
			},
		})
	}

	DB.Config.Bearers = New
	if err := DB.DB.DB.Put([]byte(`config`), Struct2Bytes(DB.Config)); err != nil {
		panic(err)
	}

}

func randProxyUse(proxys []string) (Use *h2.ProxyAuth) {
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

func (DB *Data) SaveAvatar() {
	if err := DB.DB.DB.Put([]byte(`discord_avatar`), Struct2Bytes(DB.DiscordImageBytes)); err != nil {
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
