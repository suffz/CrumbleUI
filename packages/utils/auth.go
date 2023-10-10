package utils

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"main/packages/apiGO"
	"main/packages/config"

	"github.com/golang-jwt/jwt"
)

func AuthAccs() {
	grabDetails()

	if len(Con.Config.Bearers) == 0 {
		if len(Bearer.Details) == 0 {
			return
		}
	} else {
		checkifValid()
		for _, Accs := range Con.Config.Bearers {
			if Accs.NameChange {
				Bearer.Details = append(Bearer.Details, apiGO.Info{
					Bearer:      Accs.Bearer,
					AccountType: Accs.Type,
					Email:       Accs.Email,
					Password:    Accs.Password,
					Info:        apiGO.UserINFO(Accs.Info),
				})
			}
		}
	}
	fmt.Println()
}

func grabDetails() {

	var bearerinvalids []string
	var AccountsVer []string

	for _, checkbearer := range Con.Accounts {
		t, _ := jwt.Parse(checkbearer, func(t *jwt.Token) (interface{}, error) {
			return t, nil
		})
		if strings.Contains(fmt.Sprintf("%v", t), "minecraft_net") && strings.Contains(fmt.Sprintf("%v", t), "xuid:") {
			var Data apiGO.UserINFO
			var Accounttype string
			if len(Proxy.Proxys) > 0 {
				ip, port, user, pass := GetProxyStrings(Proxy.CompRand())
				Data, Accounttype = ReturnAll(checkbearer, &apiGO.ProxyMS{IP: ip, Port: port, User: user, Password: pass})
			} else {
				Data, Accounttype = ReturnAll(checkbearer, &apiGO.ProxyMS{})
			}
			if Data.Name == "" || Data.ID == "" {
				bearerinvalids = append(bearerinvalids, checkbearer)
			} else {
				Bearer.Details = append(Bearer.Details, apiGO.Info{
					Bearer:      checkbearer,
					Email:       checkbearer[0:16] + "@gmail.com",
					AccountType: Accounttype,
					Info:        Data,
				})
			}
		} else {
			AccountsVer = append(AccountsVer, checkbearer)
		}
	}

	if len(bearerinvalids) > 0 {
		Appendinvalids(bearerinvalids)
	}

	if len(AccountsVer) == 0 {
		if len(Bearer.Details) == 0 {
			return
		} else {
			return
		}
	}

	P := Payload(AccountsVer)
	AccountsVer = CheckDupes(AccountsVer)
	if Con.Config.Bearers == nil {
		P_Auth(P, false)
	} else if len(Con.Config.Bearers) < len(AccountsVer) {
		var auth []string
		check := make(map[string]bool)
		for _, Acc := range Con.Config.Bearers {
			check[Acc.Email+":"+Acc.Password] = true
		}
		for _, Accs := range AccountsVer {
			if !check[Accs] {
				auth = append(auth, Accs)
			}
		}
		P_Auth(Payload(auth), false)
	} else if len(AccountsVer) < len(Con.Config.Bearers) {
		var New []config.Bearers
		for _, Accs := range AccountsVer {
			for _, num := range Con.Config.Bearers {
				if Accs == num.Email+":"+num.Password {
					New = append(New, num)
					break
				}
			}
		}
		Con.Config.Bearers = New
	}

	Con.SaveConfig()

}

func checkifValid() {
	var reAuth []string
	var wgs sync.WaitGroup
	for _, Accs := range Con.Config.Bearers {
		if time.Now().Unix() > Accs.AuthedAt+Accs.AuthInterval {
			reAuth = append(reAuth, Accs.Email+":"+Accs.Password)
		} else {
			if Accs.NameChange {
				wgs.Add(1)
				go func(Accs config.Bearers) {
					f, _ := http.NewRequest("GET", "https://api.minecraftservices.com/minecraft/profile/name/boom/available", nil)
					f.Header.Set("Authorization", "Bearer "+Accs.Bearer)
					if j, err := http.DefaultClient.Do(f); err == nil {
						if j.StatusCode == 401 {
							reAuth = append(reAuth, Accs.Email+":"+Accs.Password)
						}
					}
					wgs.Done()
				}(Accs)
			}
		}
	}
	wgs.Wait()
	if len(reAuth) != 0 {
		P_Auth(Payload(reAuth), true)
	}
	Con.SaveConfig()

}

func CheckDupes(strs []string) []string {
	dedup := strs[:0]
	track := make(map[string]bool, len(strs))

	for _, str := range strs {
		if track[str] {
			continue
		}
		dedup = append(dedup, str)
		track[str] = true
	}

	return dedup
}

func CheckAccs() {
	for {
		time.Sleep(10 * time.Second)
		var reauth []string
		for _, acc := range Con.Config.Bearers {
			if time.Now().Unix() > acc.AuthedAt+acc.AuthInterval && acc.NameChange {
				reauth = append(reauth, acc.Email+":"+acc.Password)
			}
		}
		if len(reauth) > 0 {
			P_Auth(Payload(reauth), true)
		}
		Con.SaveConfig()

	}
}

func Payload(accounts []string) (Data []Payload_auth) {
	var use_proxy, ug int
	var f bool = true
	for _, bearer := range accounts {
		if use_proxy >= len(Proxy.Proxys) && len(Proxy.Proxys) < len(Bearer.Details) {
			break
		}

		if f {
			Data = append(Data, Payload_auth{Proxy: Proxy.Proxys[use_proxy]})
			f = false
			use_proxy++
		}
		if len(Data[ug].Accounts) != 3 {
			Data[ug].Accounts = append(Data[ug].Accounts, bearer)
		} else {
			ug++
			Data = append(Data, Payload_auth{Proxy: Proxy.Proxys[use_proxy], Accounts: []string{bearer}})
			use_proxy++
		}
	}
	return
}

func P_Auth(P []Payload_auth, reauth bool) {
	var wg sync.WaitGroup
	var Invalids []string
	var invalidproxys []string
	for _, acc_1 := range P {
		for _, p := range acc_1.Accounts {
			if acc := strings.Split(p, ":"); len(acc) > 1 {
				if len(Proxy.Proxys) > 0 && config.Con_two.Bools.UseProxyDuringAuth {
					wg.Add(1)
					go func(proxy string, acc []string) {
						ip, port, user, pass := GetProxyStrings(proxy)
						var Authed bool
						var NC_ bool
						go func() {
							for !Authed {
								time.Sleep(80 * time.Second)
								if NC_ {
									if !Authed {
										invalidproxys = append(invalidproxys, fmt.Sprintf("%v:%v:%v:%v", ip, port, user, pass))
										ip, port, user, pass = GetProxyStrings(Proxy.CompRand())
										a, NC, Inv := sendAuth(acc[0], acc[1], reauth, &apiGO.ProxyMS{IP: ip, Port: port, User: user, Password: pass})
										NC_ = NC
										Invalids = append(Invalids, Inv...)
										Authed = a
										wg.Done()
									}
								}
							}
						}()
						a, NC, Inv := sendAuth(acc[0], acc[1], reauth, &apiGO.ProxyMS{IP: ip, Port: port, User: user, Password: pass})
						NC_ = NC
						Invalids = append(Invalids, Inv...)
						Authed = a
						wg.Done()
					}(acc_1.Proxy, acc)
				} else {
					_, _, Inv := sendAuth(acc[0], acc[1], reauth, nil)
					Invalids = append(Invalids, Inv...)
				}
			}
		}
	}
	wg.Wait()
	if len(Invalids) != 0 {
		Appendinvalids(Invalids)
		scanner := bufio.NewScanner(strings.NewReader(strings.Join(Invalids, "\n")))
		for scanner.Scan() {
			for i, acc := range Con.Config.Bearers {
				if strings.EqualFold(acc.Email, strings.Split(scanner.Text(), ":")[0]) {
					Con.Config.Bearers[i].NameChange = false
					Con.SaveConfig()

					break
				}
			}
		}
	}
	if len(invalidproxys) != 0 {
		if body, err := os.ReadFile("data/proxys.txt"); err == nil {
			strings.ReplaceAll(string(body), strings.Join(invalidproxys, "\n"), "")
			fmt.Println(strings.ReplaceAll(string(body), strings.Join(invalidproxys, "\n"), ""))
		}
	}
}

func sendAuth(email, password string, reauth bool, proxy *apiGO.ProxyMS) (Authed bool, NC bool, Invalids []string) {
	info := apiGO.MS_authentication(email, password, proxy)
	if info.Error != "" {
		Invalids = append(Invalids, email+":"+password)
	} else if info.Bearer != "" {
		fmt.Println(proxy)
		var p string
		if proxy != nil {
			p = proxy.IP + ":" + proxy.Port + ":" + proxy.User + ":" + proxy.Password
		}

		if IsChangeable(p, info.Bearer) {
			Authed = true
			NC = true
			if reauth {
				for point, bf := range Con.Config.Bearers {
					if strings.EqualFold(bf.Email, info.Email) {
						Con.Config.Bearers[point] = config.Bearers{
							Bearer:       info.Bearer,
							NameChange:   true,
							Type:         info.AccountType,
							Password:     info.Password,
							Email:        info.Email,
							AuthedAt:     time.Now().Unix(),
							AuthInterval: 54000,
							Info: config.UserINFO{
								ID:   info.Info.ID,
								Name: info.Info.Name,
							},
						}
						break
					}
				}
				for i, Bearers := range Bearer.Details {
					if strings.EqualFold(Bearers.Email, info.Email) {
						Bearer.Details[i] = info
						break
					}
				}
				var Found bool
			E1:
				for i, accs := range Accs["Giftcard"] {
					for e, b := range accs.Accs {
						if strings.EqualFold(b.Email, info.Email) {
							Accs["Giftcard"][i].Accs[e] = info
							Found = true
							break E1
						}
					}
				}
				if !Found {
				E2:
					for i, accs := range Accs["Microsoft"] {
						for e, b := range accs.Accs {
							if strings.EqualFold(b.Email, info.Email) {
								Accs["Microsoft"][i].Accs[e] = info
								Found = true
								break E2
							}
						}
					}
				}
			} else {
				Con.Config.Bearers = append(Con.Config.Bearers, config.Bearers{
					Bearer:       info.Bearer,
					AuthInterval: 54000,
					AuthedAt:     time.Now().Unix(),
					Type:         info.AccountType,
					Email:        info.Email,
					Password:     info.Password,
					NameChange:   true,
					Info: config.UserINFO{
						ID:   info.Info.ID,
						Name: info.Info.Name,
					},
				})
			}
		} else {
			NC = false
			for point, bf := range Con.Config.Bearers {
				if strings.EqualFold(bf.Email, info.Email) {
					Con.Config.Bearers[point] = config.Bearers{
						Type:         info.AccountType,
						Bearer:       info.Bearer,
						NameChange:   false,
						Password:     info.Password,
						Email:        info.Email,
						AuthedAt:     time.Now().Unix(),
						AuthInterval: 54000,
						Info:         config.UserINFO(info.Info),
					}
					break
				}
			}
			for i, Bearers := range Bearer.Details {
				if strings.EqualFold(Bearers.Email, info.Email) {
					Bearer.Details[i] = info
					break
				}
			}
			var Found bool
		E13:
			for i, accs := range Accs["Giftcard"] {
				for e, b := range accs.Accs {
					if strings.EqualFold(b.Email, info.Email) {
						Accs["Giftcard"][i].Accs[e] = info
						Found = true
						break E13
					}
				}
			}
			if !Found {
			E23:
				for i, accs := range Accs["Microsoft"] {
					for e, b := range accs.Accs {
						if strings.EqualFold(b.Email, info.Email) {
							Accs["Microsoft"][i].Accs[e] = info
							Found = true
							break E23
						}
					}
				}
			}
			Invalids = append(Invalids, email+":"+password)
		}
	}
	return
}
