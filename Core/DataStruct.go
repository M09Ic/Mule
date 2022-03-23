package Core

import (
	"Mule/utils"
	"net/http"
)

type Options struct {
	Thread      int
	Timeout     int
	Headers     []HTTPHeader
	Dictionary  []string
	PoolSize    int
	DirRoot     string
	Range       string
	Target      []string
	Cookie      string
	Prefix      string
	Method      string
	Mod         string
	TargetRange string
	JsFinder    bool
	Nolog       bool
	NoUpdate    bool
	Follow      bool
	Nobanner    bool
	Transport   *http.Transport
	utils.AutoDict
}

type ReqRes struct {
	StatusCode int `json:"StatusCode"`
	Header     http.Header
	Body       []byte
	Length     int64 `json:"Length"`
}

type WildCard struct {
	StatusCode int
	Location   string
	Body       []byte
	Length     int64
	Type       int
}

type WafCk struct {
	WafName string
	Alive   bool
}
