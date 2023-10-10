package main

import (
	"main/packages/apiGO"
	"main/packages/config"
	"main/packages/utils"
	"main/packages/utils/followbot"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/faiface/beep"
	"github.com/suffz/Youtube"
)

var D = []string{}

func Len() int {
	return len(D)
}

type TempDataForCanvas struct {
	window fyne.Window
}

func (T *TempDataForCanvas) Canvas() fyne.CanvasObject {
	but := &widget.Button{
		Text:      "template",
		Icon:      widget.NewIcon(GlobalResource).Resource,
		Alignment: widget.ButtonAlignLeading,
	}

	but.OnTapped = func() {
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
						T.window.Clipboard().SetContent(but.Text)
					}
					if Delete {
						var N []string
						for _, N_ := range D {
							if strings.Replace(N_, "<send_admin_png>", "", 1) != but.Text {
								N = append(N, N_)
							}
						}
						D = N
						list.Refresh()
					}
				}
			}, T.window),
		)
	}

	return but
}

func Update(i widget.ListItemID, o fyne.CanvasObject) {
	data := o.(*widget.Button)
	switch true {
	case strings.Contains(D[i], "<send_admin_png>"):
		data.SetText(strings.Replace(D[i], "<send_admin_png>", "", 1))
		data.Icon = widget.NewIcon(ll).Resource
	default:
		data.SetText(D[i])
	}
}

func ResizeAndShowDialog(Dialog dialog.Dialog) {
	Dialog.Resize(fyne.NewSize(400, 120))
	Dialog.Show()
}

func AddMessage(Content string, list *widget.List) {
	D = append(D, Content)
	list.Refresh()
	list.ScrollToBottom()
}

func TempCalc(interval, accamt int) time.Duration {
	return time.Duration(interval/accamt) * time.Millisecond
}

var CURRSONG string

func Youtube_INIT() {
	if config.Con_two.Settings.Youtube != "" {
		if strings.Contains(config.Con_two.Settings.Youtube, "playlist?list=") {
			Reqs := Youtube.Playlist(config.Con_two.Settings.Youtube)
			go func() {

				var Data []followbot.YTVIDEO
				audio, _ := Youtube.Video(Reqs[0], []string{}, []string{Youtube.AudioLow, Youtube.AudioMedium}, false)
				for audio.URL == "" {
					audio, _ = Youtube.Video(Reqs[0], []string{}, []string{Youtube.AudioLow, Youtube.AudioMedium}, false)
				}
				streamer, format := Youtube.Ffmpeg(audio.Download())
				Data = append(Data, followbot.YTVIDEO{
					S:        streamer,
					F:        format,
					Vid:      Reqs[0],
					Streamer: audio,
				})

				go func() {
					for _, req := range Reqs[1:] {
						audio, _ := Youtube.Video(req, []string{}, []string{Youtube.AudioLow, Youtube.AudioMedium}, false)
						for audio.URL == "" {
							audio, _ = Youtube.Video(Reqs[0], []string{}, []string{Youtube.AudioLow, Youtube.AudioMedium}, false)
						}
						streamer, format := Youtube.Ffmpeg(audio.Download())
						Data = append(Data, followbot.YTVIDEO{
							S:        streamer,
							F:        format,
							Vid:      req,
							Streamer: audio,
						})
					}
				}()

				for {
					for i, req := range Data {
						if !req.PlayedAlr {
							Data[i].PlayedAlr = true
							CURRSONG = req.Vid.Title
							s := beep.NewBuffer(req.F)
							s.Append(req.S)
							req.Streamer.PlayWithStream(s)
						}
					}
				}
			}()
		} else {
			ID := Youtube.YoutubeURL(config.Con_two.Settings.Youtube)
			YT, _ := Youtube.Video(Youtube.Youtube{
				ID: ID,
			}, []string{}, []string{Youtube.AudioLow, Youtube.AudioMedium}, false)
			CURRSONG = YT.Config.Title
			go func() {
				streamer, format := Youtube.Ffmpeg(YT.Download())
				s := beep.NewBuffer(format)
				s.Append(streamer)
				for {
					YT.PlayWithStream(s)
				}
			}()
		}
	}
}

func setUPINITACCS() {
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
				var am int = config.Con_two.Settings.AccountsPerMfa
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
				var am int = config.Con_two.Settings.AccountsPerGc
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
