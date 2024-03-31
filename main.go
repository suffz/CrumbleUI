package main

import (
	"crypto/tls"
	"embed"
	"errors"
	"fmt"
	"main/packages/apiGO"
	"main/packages/utils"
	"main/packages/utils/followbot"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/gen2brain/beeep"
	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/*
var assets embed.FS

func init() {

	utils.Roots.AppendCertsFromPEM(utils.ProxyByte)
	if _, err := os.Stat("data"); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll("data", os.ModePerm)
	}
	utils.Con.LoadState()

	sort.Slice(utils.Con.Tasks, func(i, j int) bool {
		return utils.Con.Tasks[i].Droptime_Start < utils.Con.Tasks[j].Droptime_Start
	})

	utils.Con.Bools.UseProxyDuringAuth = true
	var All []string
	for _, prox := range utils.Con.Proxys {
		if prox.User != "" {
			All = append(All, fmt.Sprintf("%v:%v:%v:%v", prox.IP, prox.Port, prox.User, prox.Password))
		} else {
			All = append(All, fmt.Sprintf("%v:%v", prox.IP, prox.Port))
		}
	}
	utils.Proxy.GetProxys(true, All)
	utils.Proxy.Setup()

	utils.AuthAccs(true, []string{})
	go utils.CheckAccs()

	var gcamt, mfaamt int

	for _, bearer := range utils.Bearer.Details {
		if utils.Use_proxy >= len(utils.Proxy.Proxys) && len(utils.Proxy.Proxys) < len(utils.Bearer.Details) {
			break
		}
		switch bearer.AccountType {
		case "Microsoft":
			if utils.First_mfa {
				utils.Accs["Microsoft"] = []utils.Proxys_Accs{{Proxy: utils.Proxy.Proxys[utils.Use_proxy]}}
				utils.First_mfa = false
				utils.Use_proxy++
			}
			var am int = utils.Con.Settings.AccountsPerMfa
			if am == 0 {
				am = 1
			}
			if len(utils.Accs["Microsoft"][utils.Use_mfa].Accs) != am {
				utils.Accs["Microsoft"][utils.Use_mfa].Accs = append(utils.Accs["Microsoft"][utils.Use_mfa].Accs, bearer)
				utils.Accamt++
				mfaamt++
			} else {
				utils.Use_mfa++
				utils.Accamt++
				mfaamt++
				utils.Accs["Microsoft"] = append(utils.Accs["Microsoft"], utils.Proxys_Accs{Proxy: utils.Proxy.Proxys[utils.Use_proxy], Accs: []apiGO.Info{bearer}})
				utils.Use_proxy++
			}
		case "Giftcard":
			if utils.First_gc {
				utils.Accs["Giftcard"] = []utils.Proxys_Accs{{Proxy: utils.Proxy.Proxys[utils.Use_proxy]}}
				utils.First_gc = false
				utils.Use_proxy++
			}
			var am int = utils.Con.Settings.AccountsPerGc
			if am == 0 {
				am = 1
			}
			if len(utils.Accs["Giftcard"][utils.Use_gc].Accs) != am {
				utils.Accs["Giftcard"][utils.Use_gc].Accs = append(utils.Accs["Giftcard"][utils.Use_gc].Accs, bearer)
				utils.Accamt++
				gcamt++
			} else {
				utils.Use_gc++
				utils.Accamt++
				gcamt++
				utils.Accs["Giftcard"] = append(utils.Accs["Giftcard"], utils.Proxys_Accs{Proxy: utils.Proxy.Proxys[utils.Use_proxy], Accs: []apiGO.Info{bearer}})
				utils.Use_proxy++
			}
		}
	}

	if gcamt == 0 {
		gcamt = 1
	}
	if mfaamt == 0 {
		mfaamt = 1
	}

	if len(utils.Con.Tasks) != 0 {
		for _, task := range utils.Con.Tasks {
			if task.Active {
				TasksActive[task.Name] = task
			}
		}
	}

	go detectTasks()
	go readyTasks()
}

func main() {

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		//Frameless: true,
		Title:  "Crumble 🔥",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			DisableWindowIcon:    true,
			BackdropType:         windows.Mica,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

var TasksActive map[string]utils.Tasks = make(map[string]utils.Tasks)
var CurrentRunning map[string]utils.Tasks = make(map[string]utils.Tasks)
var TaskLogs map[string][]*apiGO.Details = make(map[string][]*apiGO.Details)

func detectTasks() {
	for {
		if len(utils.Con.Tasks) > 0 {
			for _, task := range utils.Con.Tasks {
				if _, ok := TasksActive[task.Name]; !ok {
					TasksActive[task.Name] = task
				}
			}
		}
		time.Sleep(time.Second * 5)
	}
}

func readyTasks() {
	for {
	Exit:
		for _, task := range TasksActive {

			var start, end = task.Droptime_Start, task.Droptime_End
			drop := time.Unix(int64(start), 0)

			if time.Now().Unix() < drop.Unix() {

				CurrentRunning[task.Name] = task

				for time.Now().Before(drop) && TasksActive[task.Name].Active {
					if !CurrentRunning[task.Name].Active {
						delete(CurrentRunning, task.Name)
						break Exit
					}
					time.Sleep(time.Second * 1)
				}

				delete(CurrentRunning, task.Name)
			}

			go func() {
			Exit:
				for TasksActive[task.Name].Active {
					if utils.IsAvailable(task.Name) {

						task := TasksActive[task.Name]
						task.Active = false
						TasksActive[task.Name] = task

						break Exit
					}
					if start != 0 && end != 0 && time.Now().After(time.Unix(int64(end), 0)) {
						task := TasksActive[task.Name]
						task.Active = false
						TasksActive[task.Name] = task

						break Exit
					}
					time.Sleep(10 * time.Second)
				}
			}()
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
					Spread: func(interval, accamt int) time.Duration {
						return time.Duration(interval/accamt) * time.Millisecond
					}(utils.Con.Settings.SleepAmtPerGc, utils.Accamt),
				})
			}

			for _, Acc := range utils.Accs["Microsoft"] {
				Payloads = append(Payloads, Proxys{
					Accounts: Acc.Accs,
					Proxy:    Acc.Proxy,
					Spread: func(interval, accamt int) time.Duration {
						return time.Duration(interval/accamt) * time.Millisecond
					}(utils.Con.Settings.SleepAmtPerMfa, utils.Accamt),
				})
			}

			for TasksActive[task.Name].Active {
				for _, c := range Payloads {
					for _, accs := range c.Accounts {
						go func(Config apiGO.Info, c Proxys) {
							var wg sync.WaitGroup
							if P, ok, _ := utils.Connect(c.Proxy); ok {
								for i := 0; i < map[string]int{"Giftcard": utils.Con.Settings.GC_ReqAmt, "Microsoft": utils.Con.Settings.MFA_ReqAmt}[Config.AccountType]; i++ {
									wg.Add(1)
									go func() {
										Req := (&apiGO.Details{ResponseDetails: apiGO.SocketSending(P, utils.ReturnPayload(Config.AccountType, Config.Bearer, task.Name)), Bearer: Config.Bearer, Email: Config.Email, Type: Config.AccountType})

										TaskLogs[task.Name] = append(TaskLogs[task.Name], Req)

										switch Req.ResponseDetails.StatusCode {
										case "200":

											beeep.Notify("Success!", fmt.Sprintf(`You claimed %v!`, task.Name), "")

											if utils.Con.SkinChange.Link != "" {
												apiGO.ChangeSkin(apiGO.JsonValue(utils.Con.SkinChange), Config.Bearer)
											}
											NMC := utils.Namemc_key(Config.Bearer)
											if utils.Con.NMC.UseNMC {
												followbot.Claim_NAMEMC(NMC)
												followbot.SendFollow(task.Name)
											}
										}

										wg.Done()
									}()
									time.Sleep(func(interval, accamt int) time.Duration {
										return time.Duration(interval/accamt) * time.Millisecond
									}(map[string]int{"Giftcard": utils.Con.Settings.SleepAmtPerGc, "Microsoft": utils.Con.Settings.SleepAmtPerMfa}[Config.AccountType], utils.Accamt*2))
								}
							}
							wg.Wait()
						}(accs, c)
					}
					time.Sleep(map[bool]time.Duration{true: time.Duration(utils.Con.Settings.Spread) * time.Millisecond, false: c.Spread}[utils.Con.Settings.UseCustomSpread])
				}
			}
			delete(TasksActive, task.Name)
		}
	}
}
