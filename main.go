package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"main/packages/StrCmd"
	"main/packages/apiGO"
	"main/packages/config"
	"main/packages/utils"
	"main/packages/utils/followbot"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/image/webp"
)

func init() {

	if _, err := os.Stat("data"); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll("data", os.ModePerm)
		os.Create("data/config.json")
		os.Mkdir("data/yt", os.ModePerm)
	}

	if files, err := os.ReadDir("data/yt"); err == nil {
		for _, file := range files {
			os.Remove("data/yt/" + file.Name())
			// please clear ur recycle bin when needed.
		}
	}

	config.Con_two.LoadState()
	utils.Con = config.C()

	go Youtube_INIT()

	utils.Roots.AppendCertsFromPEM(utils.ProxyByte)

	var P []string
	for _, p := range utils.Con.Proxys {
		var addon string
		if p.User == "" {
			addon = p.Ip + ":" + p.Port
		} else {
			addon = p.Ip + ":" + p.Port + ":" + p.User + ":" + p.Password
		}
		P = append(P, addon)
	}

	utils.Proxy.GetProxys(true, P)
	utils.Proxy.Setup()
	utils.AuthAccs()

	go utils.CheckAccs()
	setUPINITACCS()

}

var GlobalResource fyne.Resource
var ll, _ = fyne.LoadResourceFromURLString("https://avatars.githubusercontent.com/u/84757238?v=4")
var list *widget.List
var toolbar *widget.Toolbar

func main() {

	app, msg, input := app.New(), container.NewVBox(), widget.NewEntry()
	window := app.NewWindow(`C:\tmp\db\genocide 😈`)

	list = widget.NewList(
		Len,
		(&TempDataForCanvas{window: window}).Canvas,
		Update)

	app.Settings().SetTheme(theme.DefaultTheme())
	Scroll := container.NewScroll(msg)

	Scroll.SetMinSize(fyne.Size{Width: 3, Height: 5})

	input.PlaceHolder = "Type Help!"

	toolbar = widget.NewToolbar(
		widget.NewToolbarAction(ll, func() {
			updateDiscordFunc(window)
		}),
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			Multi := widget.NewMultiLineEntry()
			Multi.SetPlaceHolder("email:password")

			var accs string

			for _, data := range utils.Con.Config.Bearers {
				accs += data.Email + ":" + data.Password + "\n"
			}

			Multi.SetText(accs)
			Multi.SetMinRowsVisible(13)

			ResizeAndShowDialog(dialog.NewForm("Accounts", "OK", "Close", []*widget.FormItem{
				widget.NewFormItem("", Multi),
			}, func(b bool) {
				if b {

					Multi.Refresh()

					accs := []string{}
					scanner := bufio.NewScanner(strings.NewReader(Multi.Text))
					for scanner.Scan() {
						if content := strings.Split(scanner.Text(), ":"); len(content) > 1 {
							accs = append(accs, scanner.Text())
						}
					}

					var removedupes map[string]string = make(map[string]string)

					for _, a := range accs {
						removedupes[a] = ""
					}

					var convertedtype []apiGO.Info

					for _, Temp := range utils.Con.Config.Bearers {
						if _, ok := removedupes[Temp.Email+":"+Temp.Password]; ok {
							convertedtype = append(convertedtype, apiGO.Info{
								Bearer:      Temp.Bearer,
								Email:       Temp.Email,
								Password:    Temp.Password,
								AccountType: Temp.Type,
								Info: apiGO.UserINFO{
									Name: Temp.Info.Name,
									ID:   Temp.Info.ID,
								},
							})
						}
					}

					var data int
					for range removedupes {
						data++
					}

					if len(convertedtype) < len(utils.Con.Config.Bearers) {
						SaveAccount(convertedtype, utils.Proxy.Proxys, &utils.Con)
						return
					} else if data >= len(utils.Con.Config.Bearers) {
						var accs_to_auth []apiGO.Info
						for _, a := range accs {
							e := strings.Split(a, ":")
							data := apiGO.MS_authentication(e[0], e[1], (*apiGO.ProxyMS)(utils.Proxy.RandomProxyWithStruct()))
							if data.AccountType == "Microsoft" && data.Bearer != "" && data.Error == "" {
								AddMessage(fmt.Sprintf("<send_admin_png><Crumble> Succesfully authed %v", data.Email), list)
								accs_to_auth = append(accs_to_auth, data)
							} else {
								AddMessage(fmt.Sprintf("<send_admin_png><Crumble> Unable to auth %v - %v", data.Email, data.Error), list)
							}
						}
						SaveAccount(accs_to_auth, utils.Proxy.Proxys, &utils.Con)
					}

					utils.Accs["Giftcard"] = []utils.Proxys_Accs{}
					utils.Accs["Microsoft"] = []utils.Proxys_Accs{}
					setUPINITACCS()
				}
			}, window))
		}), // accs
		widget.NewToolbarAction(theme.StorageIcon(), func() {
			Multi := widget.NewMultiLineEntry()
			Multi.SetPlaceHolder("ip:port:user:password")
			var accs string
			for _, data := range utils.Con.Proxys {
				accs += data.Ip + ":" + data.Port + ":" + data.User + ":" + data.Password + "\n"
			}

			Multi.SetText(accs)
			Multi.SetMinRowsVisible(13)

			ResizeAndShowDialog(dialog.NewForm("Proxys", "OK", "Close", []*widget.FormItem{
				widget.NewFormItem("", Multi),
			}, func(b bool) {
				if b {
					Multi.Refresh()
					accs := []config.Proxys{}
					var raw []string
					scanner := bufio.NewScanner(strings.NewReader(Multi.Text))
					for scanner.Scan() {
						data := scanner.Text()
						if content := strings.Split(data, ":"); len(content) == 4 || len(content) == 2 {
							raw = append(raw, data)
							accs = append(accs, config.ProxyData(scanner.Text()))
						}
					}

					utils.Con.Proxys = accs
					utils.Proxy.Proxys = raw
					utils.Proxy.Setup()

					utils.Accs["Giftcard"] = []utils.Proxys_Accs{}
					utils.Accs["Microsoft"] = []utils.Proxys_Accs{}
					setUPINITACCS()

					utils.Con.SaveProxys(utils.Con.Proxys)
				}
			}, window))
		}), // proxys
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			//wip
		}), // utils.Con
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.ContentRemoveIcon(), func() {
			window.Close()
		}),
	)

	/*
		container.NewBorder(
			nil, nil,
			widget.NewIcon(theme.AccountIcon()),
			nil,
			widget.NewRichTextFromMarkdown("http2\n\nhi!"),
		)
	*/

	window.SetContent(
		container.NewBorder(
			toolbar,
			nil,
			nil,
			nil,
			container.NewVScroll(
				container.NewBorder(
					nil,
					container.NewBorder(
						nil,
						nil,
						nil,
						widget.NewButton(
							"Clear",
							func() {
								ResizeAndShowDialog(
									dialog.NewForm(
										"",
										"YES",
										"NO",
										[]*widget.FormItem{
											widget.NewFormItem("", widget.NewTextGridFromString("Are you sure you want to clear ALL logs?")),
										}, func(b bool) {
											if b {
												D = []string{}
												list.Refresh()
											}
										}, window),
								)
							}),
						input),
					nil,
					nil, Scroll, list))))

	window.Resize(fyne.NewSize(1080, 720))
	window.CenterOnScreen()

	var cl, Changed, Running bool

	appdata := StrCmd.App{
		DontUseBuiltinHelpCmd: true,
		Version:               "v1.1.0",
		AppDescription:        "Crumble is a open source minecraft turbo!",
		Commands: map[string]StrCmd.Command{
			"clear": {
				Action: func() {
					D = []string{}
					list.Refresh()
				},
			},
			"help": {
				Action: func() {
					AddMessage(fmt.Sprintf("<send_admin_png>Commands:"), list)
					AddMessage(fmt.Sprintf("<send_admin_png>snipe -u <name>"), list)
				},
			},
			"snipe": {
				Description: "Snipe a minecraft name!",
				Args: []string{
					"-u", "-start", "-end",
				},
				Subcommand: map[string]StrCmd.SubCmd{
					"stop": {
						Action: func() {
							if Running {
								cl = true
								Changed = true
								Running = false
							}
						},
					},
				},
				Action: func() {
					go func() {
						if len(utils.Bearer.Details) != 0 {
							Running = true
							name, EmailClaimed := StrCmd.String("-u"), ""
							var start, end int64 = int64(StrCmd.Int("-start")), int64(StrCmd.Int("-end"))
							if config.Con_two.NMC.UseNMC {
								start, end, _, _ = followbot.GetDroptimes(name)
							} else if !config.Con_two.NMC.UseNMC || start == 0 || end == 0 {
								if !config.Con_two.Settings.AskForUnixPrompt {
									if resp, err := http.Get("https://namemc.info/data/info/" + name); err == nil && resp.StatusCode == 200 {
										var Data utils.NInfo
										json.Unmarshal([]byte(apiGO.ReturnJustString(io.ReadAll(resp.Body))), &Data)
										start = Data.Data.StartDate.Unix()
										end = Data.Data.EndDate.Unix()
									}
								}
							}

							var namemc followbot.NameRequest

							if config.Con_two.NMC.UseNMC {
								if start != 0 || end != 0 {
									namemc = followbot.Info(name)
									go func() {
										for !cl || time.Now().Unix() < namemc.Start_Unix {
											time.Sleep(time.Minute)
											namemc = followbot.Info(name)
										}
									}()
								}
							}

							drop := time.Unix(int64(start), 0)
							var Data string = fmt.Sprintf("<send_admin_png>[%v] %v", name, time.Until(drop).Round(time.Second))
							AddMessage(Data, list)

							for time.Now().Before(drop) && !cl && !Changed {
								if config.Con_two.NMC.UseNMC {
									if start != 0 || end != 0 {
										AddMessage((fmt.Sprintf("<send_admin_png>[%v] %v | Views - %v | Status - %v                \r", name, time.Until(drop).Round(time.Second), namemc.Searches, namemc.Status)), list)
									}
								} else {
									var Temp string
									for i, info := range D {
										if strings.EqualFold(info, Data) {
											Temp = fmt.Sprintf("<send_admin_png>[%v] %v", name, time.Until(drop).Round(time.Second))
											D[i] = Temp
										}
									}
									list.Refresh()
									Data = Temp
								}
								time.Sleep(time.Second * 1)
							}

							go func() {
							Exit:
								for {
									if utils.IsAvailable(name) {
										Changed = true
										cl = true
										break Exit
									}
									if start != 0 && end != 0 && time.Now().After(time.Unix(int64(end), 0)) {
										Changed = true
										cl = true
										break Exit
									}
									time.Sleep(10 * time.Second)
								}
							}()
							go func() {
								type Proxys struct {
									Conn     *tls.Conn
									Accounts []apiGO.Info
									Proxy    string
									Spread   time.Duration
								}
								var (
									Payloads []Proxys
								)

								for _, Acc := range utils.Accs["Giftcard"] {
									Payloads = append(Payloads, Proxys{
										Accounts: Acc.Accs,
										Proxy:    Acc.Proxy,
										Spread:   TempCalc(config.Con_two.Settings.SleepAmtPerGc, utils.Accamt),
									})
								}

								for _, Acc := range utils.Accs["Microsoft"] {
									Payloads = append(Payloads, Proxys{
										Accounts: Acc.Accs,
										Proxy:    Acc.Proxy,
										Spread:   TempCalc(config.Con_two.Settings.SleepAmtPerMfa, utils.Accamt),
									})
								}

								for !cl || !Changed {
									for _, c := range Payloads {
										for _, accs := range c.Accounts {
											go func(Config apiGO.Info, c Proxys) {
												if P, ok, _ := utils.Connect(c.Proxy); ok {
													var wg sync.WaitGroup
													for i := 0; i < map[string]int{"Giftcard": config.Con_two.Settings.GC_ReqAmt, "Microsoft": config.Con_two.Settings.MFA_ReqAmt}[Config.AccountType]; i++ {
														wg.Add(1)
														go func() {
															if Req := (&apiGO.Details{ResponseDetails: apiGO.SocketSending(P, utils.ReturnPayload(Config.AccountType, Config.Bearer, name)), Bearer: Config.Bearer, Email: Config.Email, Type: Config.AccountType}); Req.ResponseDetails.StatusCode == "200" {
																if config.Con_two.SkinChange.Link != "" {
																	apiGO.ChangeSkin(apiGO.JsonValue(config.Con_two.SkinChange), Config.Bearer)
																}
																NMC := utils.Namemc_key(Config.Bearer)
																if config.Con_two.NMC.UseNMC {
																	followbot.Claim_NAMEMC(NMC)
																	followbot.SendFollow(name)
																}
																EmailClaimed = fmt.Sprint((fmt.Sprintf("OK %v claimed %v @ %v -> ~ %v ~\n", Config.Email, name, time.Now().Format("05.0000"), NMC)))
																cl = true
															} else {
																AddMessage(fmt.Sprintf(`<send_admin_png>x <%v> [%v] %v -> %v`, time.Now().Format("05.0000"), Req.ResponseDetails.StatusCode, name, utils.HashEmailClean(Config.Email)), list)
															}
															wg.Done()
														}()
													}
													wg.Wait()
												}
											}(accs, c)
										}
										time.Sleep(map[bool]time.Duration{true: time.Duration(config.Con_two.Settings.Spread) * time.Millisecond, false: c.Spread}[config.Con_two.Settings.UseCustomSpread])
									}
								}

								cl = false
								Changed = false

							}()

							go func() {
								for {
									if cl || Changed {
										if EmailClaimed == "" {
											EmailClaimed = ("No account has sniped the name.")
										}
										AddMessage("<send_admin_png>"+EmailClaimed, list)
										break
									}
									time.Sleep(1 * time.Second)
								}
							}()

						} else {
							if len(utils.Con.Config.Bearers) == 0 && len(utils.Bearer.Details) == 0 {
								return
							}
						}
					}()
				},
			},
		},
	}

	input.OnSubmitted = func(s string) {
		AddMessage("> "+input.Text, list)
		input.SetText("")
		msg.Refresh()
		if err := appdata.Input(s); err != nil {
			AddMessage("<send_admin_png>"+err.Error(), list)
		}
	}

	l := app.Lifecycle()

	l.SetOnStarted(func() {
		if config.Con_two.Bools.FirstUse {
			config.Con_two.Bools.FirstUse = false
			Multi := widget.NewEntry()
			Multi.SetMinRowsVisible(13)
			ResizeAndShowDialog(dialog.NewForm("Discord ID", "YES", "NO", []*widget.FormItem{
				widget.NewFormItem("", Multi),
			}, func(b bool) {
				if b {
					id_ := Multi.Text
					config.Con_two.DiscordID = id_
					if resp, err := http.Get("https://namemc.info/data/discord/" + id_); err == nil && resp.StatusCode == 200 {
						body, _ := io.ReadAll(resp.Body)
						var D config.NMC_A
						json.Unmarshal(body, &D)

						config.Con_two.DiscordURL = D.Data.Avatar
						config.Con_two.DiscordIGN = D.Data.Username

						config.Con_two.SaveConfig()

						ff, _ := fyne.LoadResourceFromURLString(strings.ReplaceAll(config.Con_two.DiscordURL, `"`, ""))
						toolbar.Items[0] = widget.NewToolbarAction(ff, func() {
							updateDiscordFunc(window)
						})

						GlobalResource = ff

						toolbar.BaseWidget.Refresh()
						toolbar.Refresh()

						AddMessage(fmt.Sprintf("<send_admin_png>Welcome %v!", config.Con_two.DiscordIGN), list)

					}
				}
			}, window))
		} else {
			if config.Con_two.DiscordURL != "" {

				ff, _ := fyne.LoadResourceFromURLString(strings.ReplaceAll(config.Con_two.DiscordURL, `"`, ""))
				toolbar.Items[0] = widget.NewToolbarAction(ff, func() {
					updateDiscordFunc(window)
				})

				GlobalResource = ff

				toolbar.BaseWidget.Refresh()
				toolbar.Refresh()

				AddMessage(fmt.Sprintf("<send_admin_png>Welcome %v!", config.Con_two.DiscordIGN), list)
			}
		}

	})

	window.ShowAndRun()
}

func updateDiscordFunc(window fyne.Window) {
	ID := widget.NewEntry()
	ResizeAndShowDialog(dialog.NewForm("Update Your ID", "YES", "NO", []*widget.FormItem{
		widget.NewFormItem("Discord ID", ID),
	}, func(b bool) {
		if b {
			id_ := ID.Text
			config.Con_two.DiscordID = id_
			if resp, err := http.Get("https://namemc.info/data/discord/" + id_); err == nil && resp.StatusCode == 200 {
				body, _ := io.ReadAll(resp.Body)
				var Da config.NMC_A
				json.Unmarshal(body, &Da)
				old := config.Con_two.DiscordIGN
				config.Con_two.DiscordURL = Da.Data.Avatar
				config.Con_two.DiscordIGN = Da.Data.Username
				config.Con_two.SaveConfig()

				ff, _ := fyne.LoadResourceFromURLString(strings.ReplaceAll(config.Con_two.DiscordURL, `"`, ""))
				toolbar.Items[0] = widget.NewToolbarAction(ff, func() {
					updateDiscordFunc(window)
				})

				GlobalResource = ff
				toolbar.BaseWidget.Refresh()
				toolbar.Refresh()

				var N []string
				for _, N_ := range D {
					if strings.Replace(N_, "<send_admin_png>", "", 1) != "Welcome "+old+"!" {
						N = append(N, N_)
					}
				}

				D = N
				list.Refresh()

				AddMessage("Welcome "+config.Con_two.DiscordIGN+"!", list)

			}
		}
	}, window))
}

func webP(f0 io.ReadCloser) image.Image {
	img0, _ := webp.Decode(f0)
	buff := new(bytes.Buffer)
	png.Encode(buff, img0)
	ret, _ := png.Decode(bytes.NewReader(buff.Bytes()))
	return ret

}

func SaveAccount(Temp []apiGO.Info, proxys []string, DB *config.Data) {

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
		Temp = config.CheckDupes(T)
	}

	for _, Temp := range Temp {

		var Found []int
		for i, a := range DB.Config.Bearers {
			if a.Info.Name == Temp.Info.Name {
				Found = append(Found, 1)
				DB.Config.Bearers[i].Bearer = Temp.Bearer
				DB.Config.Bearers[i].Email = Temp.Email
				DB.Config.Bearers[i].Password = Temp.Password
				DB.Config.Bearers[i].Info = config.UserINFO{
					Name: Temp.Info.Name,
					ID:   Temp.Info.ID,
				}
			}
		}
		if len(Found) == 0 {
			DB.Config.Bearers = append(DB.Config.Bearers, config.Bearers{
				Bearer:       Temp.Bearer,
				Email:        Temp.Email,
				Password:     Temp.Password,
				AuthInterval: 54000,
				AuthedAt:     time.Now().Unix(),
				Type:         Temp.AccountType,
				NameChange:   apiGO.CheckChange(Temp.Password, config.RandProxyUse(proxys)),
				Info: config.UserINFO{
					Name: Temp.Info.Name,
					ID:   Temp.Info.ID,
				},
			})
		}
	}

	var New []config.Bearers
	for _, Temp := range Temp {
		New = append(New, config.Bearers{
			Bearer:       Temp.Bearer,
			Email:        Temp.Email,
			Password:     Temp.Password,
			AuthInterval: 54000,
			AuthedAt:     time.Now().Unix(),
			Type:         Temp.AccountType,
			NameChange:   apiGO.CheckChange(Temp.Bearer, config.RandProxyUse(proxys)),
			Info: config.UserINFO{
				Name: Temp.Info.Name,
				ID:   Temp.Info.ID,
			},
		})
	}

	DB.Config.Bearers = New
	utils.Bearer.Details = Temp

	if err := DB.DB.DB.Put([]byte(`config`), config.Struct2Bytes(DB.Config)); err != nil {
		panic(err)
	}

}
