package Core

import (
	"net/http"
)

type Options struct {
	Thread     int
	Timeout    int
	Headers    []HTTPHeader
	Dictionary []string
	DirRoot    string
	Range      string
	Target     []string
	Cookie     string
	Method     string
	Mod        string
	JsFinder   bool
	Nolog      bool
	NoUpdate   bool
	Transport  *http.Transport
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
