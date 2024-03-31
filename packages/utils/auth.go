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

	"github.com/golang-jwt/jwt"
)

type Accounts struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func AuthAccs(use_accountsTxt bool, alts []string) {
	var Accounts []string
	if use_accountsTxt {
		file, _ := os.Open("data/accounts.txt")
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			pass := scanner.Text()
			Accounts = append(Accounts, pass)
		}
	} else {
		Accounts = alts
	}

	bearers := grabDetails(use_accountsTxt, Accounts)

	if !use_accountsTxt {
		checkifValid()

		for _, Accs := range bearers {
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
	} else {
		if len(Con.Bearers) == 0 {
			if len(Bearer.Details) == 0 {
				return
			}
		} else {
			checkifValid()
			for _, Accs := range Con.Bearers {
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
	}

	var remove_dupes map[string]apiGO.Info = make(map[string]apiGO.Info)

	for _, auth := range Bearer.Details {
		remove_dupes[auth.Email] = auth
	}

	var New []apiGO.Info
	for _, data := range remove_dupes {
		New = append(New, data)
	}

	Bearer.Details = New

	fmt.Println()
}

func grabDetails(use_accountsTxt bool, Accounts []string) []Bearers {

	var bearerinvalids []string

	var AccountsVer []string

	for _, emails := range Accounts {
		checkbearer := emails
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
		return []Bearers{}
	}

	AccountsVer = CheckDupes(AccountsVer)

	fmt.Println(AccountsVer)

	P := Payload(AccountsVer)

	if !use_accountsTxt {
		var auth []string
		check := make(map[string]bool)
		for _, Acc := range Con.Bearers {
			check[Acc.Email+":"+Acc.Password] = true
		}
		for _, Accs := range AccountsVer {
			if !check[Accs] {
				auth = append(auth, Accs)
			}
		}
		P_Auth(Payload(auth), false)
	} else {
		if Con.Bearers == nil {
			fmt.Println("1")
			P_Auth(P, false)
		} else if len(Con.Bearers) < len(AccountsVer) {
			fmt.Println("2")
			var auth []string
			check := make(map[string]bool)
			for _, Acc := range Con.Bearers {
				check[Acc.Email+":"+Acc.Password] = true
			}
			for _, Accs := range AccountsVer {
				if !check[Accs] {
					auth = append(auth, Accs)
				}
			}
			P_Auth(Payload(auth), false)
		} else if len(AccountsVer) < len(Con.Bearers) {
			fmt.Println("3")
			var New []Bearers
			for _, Accs := range AccountsVer {
				for _, num := range Con.Bearers {
					if Accs == num.Email+":"+num.Password {
						New = append(New, num)
						break
					}
				}
			}
			Con.Bearers = New
		}
	}

	Con.SaveConfig()
	Con.LoadState()
	return Con.Bearers
}

func checkifValid() {
	var reAuth []string
	var wgs sync.WaitGroup
	for _, Accs := range Con.Bearers {
		if time.Now().Unix() > Accs.AuthedAt+Accs.AuthInterval {
			reAuth = append(reAuth, Accs.Email+":"+Accs.Password)
		} else {
			if Accs.NameChange {
				wgs.Add(1)
				go func(Accs Bearers) {
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
	Con.LoadState()
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
		for _, acc := range Con.Bearers {
			if time.Now().Unix() > acc.AuthedAt+acc.AuthInterval && acc.NameChange {
				reauth = append(reauth, acc.Email+":"+acc.Password)
			}
		}
		if len(reauth) > 0 {
			P_Auth(Payload(reauth), true)
		}
		Con.SaveConfig()
		Con.LoadState()
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

func P_Auth(P []Payload_auth, reauth bool) []Bearers {
	var wg sync.WaitGroup
	var Invalids []string
	var invalidproxys []string
	for _, acc_1 := range P {
		for _, p := range acc_1.Accounts {
			if acc := strings.Split(p, ":"); len(acc) > 1 {
				if len(Proxy.Proxys) > 0 && Con.Bools.UseProxyDuringAuth {
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
			for i, acc := range Con.Bearers {
				if strings.EqualFold(acc.Email, strings.Split(scanner.Text(), ":")[0]) {
					Con.Bearers[i].NameChange = false
					break
				}
			}
		}
	}
	/*
		if len(invalidproxys) != 0 {
			if body, err := os.ReadFile("data/proxys.txt"); err == nil {
				strings.ReplaceAll(string(body), strings.Join(invalidproxys, "\n"), "")
				fmt.Println(strings.ReplaceAll(string(body), strings.Join(invalidproxys, "\n"), ""))
			}
		}
	*/

	Con.SaveConfig()
	Con.LoadState()
	return Con.Bearers
}

func sendAuth(email, password string, reauth bool, proxy *apiGO.ProxyMS) (Authed bool, NC bool, Invalids []string) {
	info := apiGO.MS_authentication(email, password, proxy)
	if info.Error != "" {
		Invalids = append(Invalids, email+":"+password)
	} else if info.Bearer != "" {
		var Change bool
		if proxy != nil {
			Change = IsChangeable(proxy.IP+":"+proxy.Port+":"+proxy.User+":"+proxy.Password, info.Bearer)
		} else {
			Change = IsChangeable("", info.Bearer)
		}
		if Change {
			Authed = true
			NC = true
			if reauth {
				for point, bf := range Con.Bearers {
					if strings.EqualFold(bf.Email, info.Email) {
						Con.Bearers[point] = Bearers{
							Bearer:       info.Bearer,
							NameChange:   true,
							Type:         info.AccountType,
							Password:     info.Password,
							Email:        info.Email,
							AuthedAt:     time.Now().Unix(),
							AuthInterval: 54000,
							Info: UserINFO{
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
				if !Found {
					if info.AccountType == "Microsoft" {
						if len(Accs["Microsoft"]) != 0 {
							var found bool
						E:
							for i, accs := range Accs["Microsoft"] {
								for e, b := range accs.Accs {
									if strings.EqualFold(b.Email, info.Email) {
										if reauth {
											Accs["Microsoft"][i].Accs[e] = info
										}
										found = true
										break E
									}
								}
							}
							if !found {
								if len(Accs["Microsoft"][len(Accs["Microsoft"])-1].Accs) >= Con.Settings.AccountsPerMfa {
									Accs["Microsoft"] = append(Accs["Microsoft"], []Proxys_Accs{
										{Proxy: Proxy.Proxys[Use_proxy], Accs: []apiGO.Info{info}},
									}...)
								} else {
									Accs["Microsoft"][len(Accs["Microsoft"])-1].Accs = append(Accs["Microsoft"][len(Accs["Microsoft"])-1].Accs, info)
								}
							}
						}
					} else {
						if len(Accs["Giftcard"]) != 0 {
							var found bool
						E34:
							for i, accs := range Accs["Giftcard"] {
								for e, b := range accs.Accs {
									if strings.EqualFold(b.Email, info.Email) {
										if reauth {
											Accs["Giftcard"][i].Accs[e] = info
										}
										found = true
										break E34
									}
								}
							}
							if !found {
								if len(Accs["Giftcard"][len(Accs["Giftcard"])-1].Accs) >= Con.Settings.AccountsPerMfa {
									Accs["Giftcard"] = append(Accs["Giftcard"], []Proxys_Accs{
										{Proxy: Proxy.Proxys[Use_proxy], Accs: []apiGO.Info{info}},
									}...)
								} else {
									Accs["Giftcard"][len(Accs["Giftcard"])-1].Accs = append(Accs["Giftcard"][len(Accs["Giftcard"])-1].Accs, info)
								}
							}
						}
					}
				}
			} else {

				if info.AccountType == "Microsoft" {
					if len(Accs["Microsoft"]) != 0 {
						var found bool
					E23334:
						for i, accs := range Accs["Microsoft"] {
							for e, b := range accs.Accs {
								if strings.EqualFold(b.Email, info.Email) {
									if reauth {
										Accs["Microsoft"][i].Accs[e] = info
									}
									found = true
									break E23334
								}
							}
						}
						if !found {
							if len(Accs["Microsoft"][len(Accs["Microsoft"])-1].Accs) >= Con.Settings.AccountsPerMfa {
								Accs["Microsoft"] = append(Accs["Microsoft"], []Proxys_Accs{
									{Proxy: Proxy.Proxys[Use_proxy], Accs: []apiGO.Info{info}},
								}...)
							} else {
								Accs["Microsoft"][len(Accs["Microsoft"])-1].Accs = append(Accs["Microsoft"][len(Accs["Microsoft"])-1].Accs, info)
							}
						}
					}
				} else {
					if len(Accs["Giftcard"]) != 0 {
						var found bool
					E3423:
						for i, accs := range Accs["Giftcard"] {
							for e, b := range accs.Accs {
								if strings.EqualFold(b.Email, info.Email) {
									if reauth {
										Accs["Giftcard"][i].Accs[e] = info
									}
									found = true
									break E3423
								}
							}
						}
						if !found {
							if len(Accs["Giftcard"][len(Accs["Giftcard"])-1].Accs) >= Con.Settings.AccountsPerMfa {
								Accs["Giftcard"] = append(Accs["Giftcard"], []Proxys_Accs{
									{Proxy: Proxy.Proxys[Use_proxy], Accs: []apiGO.Info{info}},
								}...)
							} else {
								Accs["Giftcard"][len(Accs["Giftcard"])-1].Accs = append(Accs["Giftcard"][len(Accs["Giftcard"])-1].Accs, info)
							}
						}
					}
				}

				Con.Bearers = append(Con.Bearers, Bearers{
					Bearer:       info.Bearer,
					AuthInterval: 54000,
					AuthedAt:     time.Now().Unix(),
					Type:         info.AccountType,
					Email:        info.Email,
					Password:     info.Password,
					NameChange:   true,
					Info: UserINFO{
						ID:   info.Info.ID,
						Name: info.Info.Name,
					},
				})
			}
		} else {
			NC = false
			for point, bf := range Con.Bearers {
				if strings.EqualFold(bf.Email, info.Email) {
					Con.Bearers[point] = Bearers{
						Type:         info.AccountType,
						Bearer:       info.Bearer,
						NameChange:   false,
						Password:     info.Password,
						Email:        info.Email,
						AuthedAt:     time.Now().Unix(),
						AuthInterval: 54000,
						Info:         UserINFO(info.Info),
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
