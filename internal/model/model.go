package model

import "time"

type Package struct {
	Name           string
	Version        string
	Size           string
	Description    string
	Installed      bool
	Upgradable     bool
	NewVersion     string
	Section        string
	Architecture   string
	SecurityUpdate bool
	Held           bool
	Pinned         bool
	Essential      bool
}

type CmdRunned struct {
	Id   uint      `json:"id"`
	Cmd  string    `json:"cmd"`
	Time time.Time `json:"time"`
}
