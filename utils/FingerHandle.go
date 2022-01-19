package utils

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

var Mmh3fingers, Md5fingers map[string]string
var Httpfingers []*Finger
var Compiled map[string][]regexp.Regexp

type FingerMapper map[string][]*Finger

type Finger struct {
	Name         string   `json:"name"`
	Protocol     string   `json:"protocol"`
	SendData_str string   `json:"send_data"`
	SendData     senddata `json:"-"`
	Vuln         string   `json:"vuln"`
	Level        int      `json:"level"`
	Defaultport  []string `json:"default_port"`
	Regexps      Regexps  `json:"regexps"`
}

type Regexps struct {
	Body   []string `json:"body"`
	MD5    []string `json:"md5"`
	Regexp []string `json:"regexp"`
	Cookie []string `json:"cookie"`
	Header []string `json:"header"`
	Vuln   []string `json:"vuln"`
}

type senddata []byte

func (fm FingerMapper) GetFingers(port string) []*Finger {
	return fm[port]
}

func (fm FingerMapper) GetOthersFingers(port string) []*Finger {
	var tmpfingers []*Finger
	for _, fingers := range fm {
		for _, finger := range fingers {
			if SliceContains(finger.Defaultport, port) {
				continue
			}
			isrepeat := false
			for _, tmpfinger := range tmpfingers {
				if finger == tmpfinger {
					isrepeat = true
				}
			}
			if !isrepeat {
				tmpfingers = append(tmpfingers, finger)
			}
		}
	}
	return tmpfingers
}

func (f *Finger) Decode() {
	if f.Protocol != "tcp" {
		return
	}

	if f.SendData_str != "" {
		f.SendData = decode(f.SendData_str)
	}
}

func decode(s string) []byte {
	var bs []byte
	if s[:4] == "b64|" {
		bs, _ = b64.StdEncoding.DecodeString(s[4:])
	} else {
		bs = []byte(s)
	}
	return bs
}

//加载指纹到全局变量
func LoadFingers(t string) []*Finger {
	var tmpfingers []*Finger

	// 根据权重排序在python脚本中已经实现
	err := json.Unmarshal([]byte(LoadConfig(t)), &tmpfingers)

	if err != nil {
		fmt.Println("[-] finger load FAIL!")
		os.Exit(0)
	}

	//初步处理tcp指纹
	for _, finger := range tmpfingers {
		finger.Decode() // 防止\xff \x00编码解码影响结果

		// 普通指纹
		for _, regstr := range finger.Regexps.Regexp {
			Compiled[finger.Name] = append(Compiled[finger.Name], CompileRegexp("(?im)"+regstr))
		}
		// 漏洞指纹,指纹名称后接 "_vuln"
		for _, regstr := range finger.Regexps.Vuln {
			Compiled[finger.Name+"_vuln"] = append(Compiled[finger.Name+"_vuln"], CompileRegexp("(?im)"+regstr))
		}
	}
	return tmpfingers
}

func LoadHashFinger() (map[string]string, map[string]string) {
	var mmh3fingers, md5fingers map[string]string
	var err error
	err = json.Unmarshal([]byte(LoadConfig("mmh3")), &mmh3fingers)
	if err != nil {
		fmt.Println("[-] mmh3 load FAIL!")
		os.Exit(0)
	}

	err = json.Unmarshal([]byte(LoadConfig("md5")), &md5fingers)
	if err != nil {
		fmt.Println("[-] mmh3 load FAIL!")
		os.Exit(0)
	}
	return mmh3fingers, md5fingers
}
