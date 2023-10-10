package config

import (
	"git.mills.io/prologic/bitcask"
)

var (
	Con_two Config_JSON
)

type NMC_A struct {
	Action string   `json:"action"`
	Desc   string   `json:"desc"`
	Code   string   `json:"code"`
	Data   Data_NMC `json:"data"`
}
type Data_NMC struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Banner   string `json:"banner"`
}

type Database struct {
	DB *bitcask.Bitcask
}

type Data struct {
	Config   Config
	Accounts []string
	Proxys   []Proxys
	DB       Database
}

type Proxys struct {
	Ip       string `json:"ip"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type Config struct {
	CF       CF          `json:"cf_tokens"`
	Bearers  []Bearers   `json:"Bearers"`
	Recovery []Succesful `json:"recovery"`
}

type Config_JSON struct {
	DiscordURL string          `json:"discord_url"`
	DiscordID  string          `json:"discord_id"`
	DiscordIGN string          `json:"discord_ign"`
	NMC        Namemc_Data     `json:"namemc_settings"`
	Settings   AccountSettings `json:"settings"`
	Bools      Bools           `json:"sniper_config"`
	SkinChange Skin            `json:"skin_config"`
}

type Succesful struct {
	Email     string
	Recovery  string
	Code_Used string
}

type Namemc_Data struct {
	UseNMC          bool   `json:"usenamemc_fordroptime_andautofollow"`
	Display         string `json:"name_to_use_for_follows"`
	Key             string `json:"namemc_email:pass"`
	NamemcLoginData NMC    `json:"namemc_login_data"`
}

type AccountSettings struct {
	Youtube          string `json:"youtube_link"`
	AskForUnixPrompt bool   `json:"ask_for_unix_prompt"`
	AccountsPerGc    int    `json:"accounts_per_gc_proxy"`
	AccountsPerMfa   int    `json:"accounts_per_mfa_proxy"`
	GC_ReqAmt        int    `json:"amt_reqs_per_gc_acc"`
	MFA_ReqAmt       int    `json:"amt_reqs_per_mfa_acc"`
	SleepAmtPerGc    int    `json:"sleep_for_gc_ms"`
	SleepAmtPerMfa   int    `json:"sleep_for_mfa_ms"`
	UseCustomSpread  bool   `json:"use_own_spread_value"`
	Spread           int64  `json:"spread_ms"`
}

type Bools struct {
	UseBypass                        bool `json:"gc_bypass"`
	UseProxyDuringAuth               bool `json:"useproxysduringauth"`
	FirstUse                         bool `json:"firstuse_IGNORETHIS"`
	DownloadedPW                     bool `json:"pwinstalled_IGNORETHIS"`
	ApplyNewRecoveryToExistingEmails bool `json:"update_recovery_on_emails_with_existing_recovery"`
}

type NMC struct {
	Token      string `json:"token"`
	LastAuthed int64  `json:"last_unix_auth_timestamp"`
}

type Bearers struct {
	Bearer               string   `json:"Bearer"`
	Email                string   `json:"Email"`
	Password             string   `json:"Password"`
	AuthInterval         int64    `json:"AuthInterval"`
	AuthedAt             int64    `json:"AuthedAt"`
	Type                 string   `json:"Type"`
	NameChange           bool     `json:"NameChange"`
	Info                 UserINFO `json:"Info"`
	NOT_ENTITLED_CHECKED bool     `json:"checked_entitled"`
}

type UserINFO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Skin struct {
	Link    string `json:"url"`
	Variant string `json:"variant"`
}

type CF struct {
	Tokens   string `json:"tokens"`
	GennedAT int64  `json:"unix_of_creation"`
}
