package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"main/packages/apiGO"
	"main/packages/utils"
	"net/http"
	"slices"
	"sort"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetRequestData(name string) string {
	resp, _ := http.Get("https://namemc.info/data/info/" + name)
	data, _ := io.ReadAll(resp.Body)
	return string(data)
}

type Logs struct {
	Name string           `json:"name"`
	Data []*apiGO.Details `json:"content"`
}

func (a *App) GetTaskLogs(name string) Logs {
	info := TaskLogs[name]
	slices.Reverse(info)
	return Logs{
		Name: name,
		Data: info,
	}
}

func (a *App) TaskAdd(t any) {

	if len(utils.Con.Proxys) == 0 {
		return
	}

	var found bool
	var data utils.Tasks

	body, _ := json.Marshal(t)
	json.Unmarshal(body, &data)

	if len(utils.Con.Tasks) != 0 {
		for _, util := range utils.Con.Tasks {
			if util.Name == data.Name {
				found = true
				break
			}
		}
	}
	if !found {
		data.Active = true

		for _, t := range CurrentRunning {
			if t.Droptime_Start > data.Droptime_Start {
				t.Active = false
				TasksActive[t.Name] = t
				CurrentRunning[t.Name] = t
			}
		}

		TasksActive[data.Name] = data
		utils.Con.Tasks = append(utils.Con.Tasks, data)

		var New map[string]utils.Tasks = make(map[string]utils.Tasks)

		sl := ToSlice(TasksActive)

		sort.Slice(sl, func(i, j int) bool {
			return sl[i].Droptime_Start < sl[j].Droptime_Start
		})

		for _, data := range sl {
			New[data.Name] = data
		}

		TasksActive = New

		sort.Slice(utils.Con.Tasks, func(i, j int) bool {
			return utils.Con.Tasks[i].Droptime_Start < utils.Con.Tasks[j].Droptime_Start
		})

		utils.Con.SaveConfig()
		utils.Con.LoadState()
	}
}

func ToSlice(m map[string]utils.Tasks) []utils.Tasks {
	cities := make([]utils.Tasks, 0, len(m))
	for k, v := range m {
		v.Name = k
		cities = append(cities, v)
	}
	return cities
}

func (a *App) TaskDelete(t any) bool {
	var data utils.Tasks
	body, _ := json.Marshal(t)
	json.Unmarshal(body, &data)
	if _, ok := TasksActive[data.Name]; ok {

		delete(TaskLogs, data.Name)

		t := TasksActive[data.Name]
		t.Active = false
		TasksActive[data.Name] = t
		var New []utils.Tasks
		if len(utils.Con.Tasks) != 0 {
			for _, util := range utils.Con.Tasks {
				if util.Name != data.Name {
					New = append(New, util)
				}
			}
		}
		utils.Con.Tasks = New
		utils.Con.SaveConfig()
		utils.Con.LoadState()
		return true
	} else {
		return false
	}
}

func (a *App) GetAllTasks() []utils.Tasks {
	return utils.Con.Tasks
}

func (a *App) AuthAccounts(info any) []apiGO.Info {
	var data []utils.Accounts
	body, _ := json.Marshal(info)
	json.Unmarshal(body, &data)

	var Accs []string
	for _, d := range data {
		Accs = append(Accs, d.Email+":"+d.Password)
	}

	utils.AuthAccs(false, Accs)

	return utils.Bearer.Details
}

func (a *App) AddProxys(info any) []utils.Proxies {
	var data []utils.Proxies
	body, _ := json.Marshal(info)
	json.Unmarshal(body, &data)

	utils.Con.Proxys = data

	utils.Proxy = apiGO.Proxys{}
	var d []string
	for _, p := range data {
		if p.Password != "" {
			d = append(d, fmt.Sprintf(`%v:%v:%v:%v`, p.IP, p.Port, p.User, p.Password))
		} else {
			d = append(d, fmt.Sprintf(`%v:%v`, p.IP, p.Port))
		}
	}

	utils.Proxy.GetProxys(true, d)
	utils.Proxy.Setup()

	if !utils.Con.Bools.UseProxyDuringAuth {
		utils.Con.Bools.UseProxyDuringAuth = true
	}

	utils.Con.SaveConfig()
	utils.Con.LoadState()

	return data
}

func (a *App) GetProxys() []string {
	var build []string

	for _, p := range utils.Con.Proxys {
		if p.Password != "" {
			build = append(build, fmt.Sprintf(`%v:%v:%v:%v`, p.IP, p.Port, p.User, p.Password))
		} else {
			build = append(build, fmt.Sprintf(`%v:%v`, p.IP, p.Port))
		}
	}
	return build
}

func (a *App) ReturnAccounts() []apiGO.Info {
	return utils.Bearer.Details
}

func (a *App) GetThreeChar() string {
	resp, _ := http.Get("https://namemc.info/data/3c")
	data, _ := io.ReadAll(resp.Body)
	return string(data)
}
