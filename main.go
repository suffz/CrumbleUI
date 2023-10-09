package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
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

	utils.Con = config.C()

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
	var use_proxy, gcamt, mfaamt int
	if len(utils.Proxy.Proxys) > 0 {
		for _, bearer := range utils.Con.Config.Bearers {
			if use_proxy >= len(utils.Proxy.Proxys) && len(utils.Proxy.Proxys) < len(utils.Bearer.Details) {
				break
			}
			switch bearer.Type {
			case "Microsoft":
				if utils.First_mfa {
					utils.Accs["Microsoft"] = []utils.Proxys_Accs{{Proxy: utils.Proxy.Proxys[use_proxy]}}
					utils.First_mfa = false
					use_proxy++
				}
				var am int = utils.Con.Config.Settings.AccountsPerMfa
				if am == 0 {
					am = 1
				}
				if len(utils.Accs["Microsoft"][utils.Use_mfa].Accs) != am {
					utils.Accs["Microsoft"][utils.Use_mfa].Accs = append(utils.Accs["Microsoft"][utils.Use_mfa].Accs, apiGO.Info{
						Bearer: bearer.Bearer,
						Email:  bearer.Email, Password: bearer.Password,
						AccountType: bearer.Type,
						Info: apiGO.UserINFO{
							ID:   bearer.Info.ID,
							Name: bearer.Info.Name,
						},
					})
					utils.Accamt++
					mfaamt++
				} else {
					utils.Use_mfa++
					utils.Accamt++
					mfaamt++
					utils.Accs["Microsoft"] = append(utils.Accs["Microsoft"], utils.Proxys_Accs{Proxy: utils.Proxy.Proxys[use_proxy], Accs: []apiGO.Info{{
						Bearer: bearer.Bearer,
						Email:  bearer.Email, Password: bearer.Password,
						AccountType: bearer.Type,
						Info: apiGO.UserINFO{
							ID:   bearer.Info.ID,
							Name: bearer.Info.Name,
						},
					}}})
					use_proxy++
				}
			case "Giftcard":
				if utils.First_gc {
					utils.Accs["Giftcard"] = []utils.Proxys_Accs{{Proxy: utils.Proxy.Proxys[use_proxy]}}
					utils.First_gc = false
					use_proxy++
				}
				var am int = utils.Con.Config.Settings.AccountsPerGc
				if am == 0 {
					am = 1
				}
				if len(utils.Accs["Giftcard"][utils.Use_gc].Accs) != am {
					utils.Accs["Giftcard"][utils.Use_gc].Accs = append(utils.Accs["Giftcard"][utils.Use_gc].Accs, apiGO.Info{
						Bearer: bearer.Bearer,
						Email:  bearer.Email, Password: bearer.Password,
						AccountType: bearer.Type,
						Info: apiGO.UserINFO{
							ID:   bearer.Info.ID,
							Name: bearer.Info.Name,
						},
					})
					utils.Accamt++
					gcamt++
				} else {
					utils.Use_gc++
					utils.Accamt++
					gcamt++
					utils.Accs["Giftcard"] = append(utils.Accs["Giftcard"], utils.Proxys_Accs{Proxy: utils.Proxy.Proxys[use_proxy], Accs: []apiGO.Info{{
						Bearer: bearer.Bearer,
						Email:  bearer.Email, Password: bearer.Password,
						AccountType: bearer.Type,
						Info: apiGO.UserINFO{
							ID:   bearer.Info.ID,
							Name: bearer.Info.Name,
						},
					}}})
					use_proxy++
				}
			}
		}

		if gcamt == 0 {
			gcamt = 1
		}
		if mfaamt == 0 {
			mfaamt = 1
		}

	}
}

func main() {

	app, msg, input := app.New(), container.NewVBox(), widget.NewEntry()
	window := app.NewWindow("Crumble 😈")

	if utils.Con.Config.Bools.FirstUse {
		utils.Con.Config.Bools.FirstUse = false
		ID := widget.NewEntry()

		id := app.NewWindow("ID 😈")

		id.Resize(fyne.NewSize(400, 120))
		id.CenterOnScreen()

		id.SetContent(&widget.Form{
			Items: []*widget.FormItem{ // we can specify items in the constructor
				{Text: "Discord ID", Widget: ID}},
			OnSubmit: func() { // optional, handle form submission
				id_ := ID.Text
				utils.Con.DiscordID = id_
				if resp, err := http.Get("https://namemc.info/data/discord/" + id_); err == nil && resp.StatusCode == 200 {
					body, _ := io.ReadAll(resp.Body)
					var D config.NMC_A
					json.Unmarshal(body, &D)
					utils.Con.DiscordImageBytes = D.Data.Avatar
					utils.Con.SaveAvatar()
				}
				id.Close()
			},
		})

		id.ShowAndRun()

	}

	ll, err := fyne.LoadResourceFromURLString(strings.ReplaceAll(utils.Con.DiscordImageBytes, `"`, ""))
	if err != nil {
		panic(err)
	}

	list := widget.NewList(
		Len,
		Canvas,
		Update)

	list.OnSelected = func(id widget.ListItemID) {
		list.Unselect(id)
		var Copy, Delete bool
		ResizeAndShowDialog(
			dialog.NewForm("Message Handler", "YES", "NO", []*widget.FormItem{
				{
					Text: "Copy Text",
					Widget: widget.NewCheck("", func(b bool) {
						Copy = b
					}),
				},
				{
					Text: "Delete Message",
					Widget: widget.NewCheck("", func(b bool) {
						Delete = b
					}),
				},
			}, func(b bool) {
				if b {
					if Copy {
						window.Clipboard().SetContent(D[id])
					}
					if Delete {
						var N []string
						for i, N_ := range D {
							if i != id {
								N = append(N, N_)
							}
						}
						D = N
						list.Refresh()
					}
				}
			}, window),
		)
	}

	app.Settings().SetTheme(theme.DefaultTheme())
	Scroll := container.NewScroll(msg)

	Scroll.SetMinSize(fyne.Size{Width: 3, Height: 5})

	input.PlaceHolder = "Type Help!"
	//window.SetMainMenu(MakeMenu(list, window))

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(fyne.NewStaticResource(ll.Name(), ll.Content()), func() {
			ID := widget.NewEntry()
			ResizeAndShowDialog(dialog.NewForm("Update Your ID", "YES", "NO", []*widget.FormItem{
				widget.NewFormItem("Discord ID", ID),
			}, func(b bool) {
				if b {
					id_ := ID.Text
					utils.Con.DiscordID = id_
					if resp, err := http.Get("https://namemc.info/data/discord/" + id_); err == nil && resp.StatusCode == 200 {
						body, _ := io.ReadAll(resp.Body)
						var D config.NMC_A
						json.Unmarshal(body, &D)
						utils.Con.DiscordImageBytes = D.Data.Avatar
						utils.Con.SaveAvatar()
					}
				}
			}, window))
		}),
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			Multi := widget.NewMultiLineEntry()
			Multi.SetPlaceHolder("email:password")

			var accs string

			for _, data := range utils.Con.Config.Bearers {
				accs += data.Email + ":" + data.Password + "\n"
			}

			Multi.SetText(accs)

			ResizeAndShowDialog(dialog.NewForm("Login", "OK", "Close", []*widget.FormItem{
				widget.NewFormItem("Accounts", Multi),
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
						utils.Con.SaveAccount(convertedtype, utils.Proxy.Proxys)
						return
					} else if data >= len(utils.Con.Config.Bearers) {
						var accs_to_auth []apiGO.Info
						for _, a := range accs {
							e := strings.Split(a, ":")
							data := apiGO.MS_authentication(e[0], e[1], (*apiGO.ProxyMS)(utils.Proxy.RandomProxyWithStruct()))
							if data.AccountType == "Microsoft" && data.Bearer != "" && data.Error == "" {
								AddMessage(fmt.Sprintf("<Crumble> Succesfully authed %v", data.Email), list)
								accs_to_auth = append(accs_to_auth, data)
							} else {
								AddMessage(fmt.Sprintf("<Crumble> Unable to auth %v - %v", data.Email, data.Error), list)
							}
						}
						utils.Con.SaveAccount(accs_to_auth, utils.Proxy.Proxys)
					}
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
			ResizeAndShowDialog(dialog.NewForm("Login", "OK", "Close", []*widget.FormItem{
				widget.NewFormItem("Accounts", Multi),
			}, func(b bool) {
				if b {
					Multi.Refresh()
					accs := []config.Proxys{}
					scanner := bufio.NewScanner(strings.NewReader(Multi.Text))
					for scanner.Scan() {
						if content := strings.Split(scanner.Text(), ":"); len(content) == 4 || len(content) == 2 {
							accs = append(accs, config.ProxyData(scanner.Text()))
						}
					}
					utils.Con.Proxys = accs
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
			"help": {},
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
					if len(utils.Bearer.Details) != 0 {
						Running = true
						name, EmailClaimed := StrCmd.String("-u"), ""
						var start, end int64 = int64(StrCmd.Int("-start")), int64(StrCmd.Int("-end"))
						if utils.Con.Config.NMC.UseNMC {
							start, end, _, _ = followbot.GetDroptimes(name)
						} else if !utils.Con.Config.NMC.UseNMC || start == 0 || end == 0 {
							if !utils.Con.Config.Settings.AskForUnixPrompt {
								if resp, err := http.Get("https://namemc.info/data/info/" + name); err == nil && resp.StatusCode == 200 {
									var Data utils.NInfo
									json.Unmarshal([]byte(apiGO.ReturnJustString(io.ReadAll(resp.Body))), &Data)
									start = Data.Data.StartDate.Unix()
									end = Data.Data.EndDate.Unix()
								}
							}
						}

						var namemc followbot.NameRequest

						if utils.Con.Config.NMC.UseNMC {
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

						for time.Now().Before(drop) {
							if utils.Con.Config.NMC.UseNMC {
								if start != 0 || end != 0 {
									AddMessage((fmt.Sprintf("[%v] %v | Views - %v | Status - %v                \r", name, time.Until(drop).Round(time.Second), namemc.Searches, namemc.Status)), list)
								}
							} else {
								AddMessage((fmt.Sprintf("[%v] %v                 \r", name, time.Until(drop).Round(time.Second))), list)
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
									Spread:   TempCalc(utils.Con.Config.Settings.SleepAmtPerGc, utils.Accamt),
								})
							}

							for _, Acc := range utils.Accs["Microsoft"] {
								Payloads = append(Payloads, Proxys{
									Accounts: Acc.Accs,
									Proxy:    Acc.Proxy,
									Spread:   TempCalc(utils.Con.Config.Settings.SleepAmtPerMfa, utils.Accamt),
								})
							}

							for !cl || !Changed {
								for _, c := range Payloads {
									for _, accs := range c.Accounts {
										go func(Config apiGO.Info, c Proxys) {
											if P, ok, _ := utils.Connect(c.Proxy); ok {
												var wg sync.WaitGroup
												for i := 0; i < map[string]int{"Giftcard": utils.Con.Config.Settings.GC_ReqAmt, "Microsoft": utils.Con.Config.Settings.MFA_ReqAmt}[Config.AccountType]; i++ {
													wg.Add(1)
													go func() {
														if Req := (&apiGO.Details{ResponseDetails: apiGO.SocketSending(P, utils.ReturnPayload(Config.AccountType, Config.Bearer, name)), Bearer: Config.Bearer, Email: Config.Email, Type: Config.AccountType}); Req.ResponseDetails.StatusCode == "200" {
															if utils.Con.Config.SkinChange.Link != "" {
																apiGO.ChangeSkin(apiGO.JsonValue(utils.Con.Config.SkinChange), Config.Bearer)
															}
															NMC := utils.Namemc_key(Config.Bearer)
															if utils.Con.Config.NMC.UseNMC {
																followbot.Claim_NAMEMC(NMC)
																followbot.SendFollow(name)
															}
															EmailClaimed = fmt.Sprint((fmt.Sprintf("✓ %v claimed %v @ %v -> ~ %v ~\n", Config.Email, name, time.Now().Format("05.0000"), NMC)))
															cl = true
														} else {
															fmt.Println(cl, Changed)
															AddMessage(fmt.Sprintf(`✗ <%v> [%v] %v -> %v`, time.Now().Format("05.0000"), Req.ResponseDetails.StatusCode, name, utils.HashEmailClean(Config.Email)), list)
														}
														wg.Done()
													}()
												}
												wg.Wait()
											}
										}(accs, c)
									}
									time.Sleep(map[bool]time.Duration{true: time.Duration(utils.Con.Config.Settings.Spread) * time.Millisecond, false: c.Spread}[utils.Con.Config.Settings.UseCustomSpread])
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
									fmt.Printf(EmailClaimed)
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
				},
			},
		},
	}

	input.OnSubmitted = func(s string) {
		input.SetText("")
		msg.Refresh()
		if err := appdata.Input(s); err != nil {
			AddMessage(err.Error(), list)
		}
	}

	window.ShowAndRun()
}

func webP(f0 io.ReadCloser) image.Image {
	img0, _ := webp.Decode(f0)
	buff := new(bytes.Buffer)
	png.Encode(buff, img0)
	ret, _ := png.Decode(bytes.NewReader(buff.Bytes()))
	return ret

}
