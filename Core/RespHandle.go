package Core

import (
	"Mule/utils"
	"fmt"
	"strings"
)

type IdentifyRes struct {
	Title      string     `json:"title"`
	Hash       string     `json:"hash"`
	Mmh3       string     `json:"mmh3"`
	SimHash    string     `json:"simhash"`
	Frameworks Frameworks `json:"Frameworks"`
}

func (result *IdentifyRes) AddFramework(f Framework) {
	result.Frameworks = append(result.Frameworks, f)
}

func (result *IdentifyRes) NoFramework() bool {
	if len(result.Frameworks) == 0 {
		return true
	}
	return false
}

type Framework struct {
	Name    string `json:"ft"`
	Version string `json:"fv"`
	IsGuess bool   `json:"fg"`
}

func (f Framework) ToString() string {
	if f.IsGuess {
		return fmt.Sprintf("*%s", f.Name)
	} else {
		if f.Version == "" {
			return fmt.Sprintf("%s", f.Name)
		} else {
			return fmt.Sprintf("%s:%s", f.Name, f.Version)
		}
	}

}

type Vuln struct {
	Name    string                 `json:"vn"`
	Payload map[string]interface{} `json:"vp"`
	Detail  map[string]interface{} `json:"vd"`
}

func (v *Vuln) GetPayload() string {
	return utils.MaptoString(v.Payload)
}

func (v *Vuln) GetDetail() string {
	return utils.MaptoString(v.Detail)
}

func (v *Vuln) ToString() string {
	s := v.Name
	if payload := v.GetPayload(); payload != "" {
		s += fmt.Sprintf(" payloads:%s", payload)
	}
	if detail := v.GetDetail(); detail != "" {
		s += fmt.Sprintf(" payloads:%s", detail)
	}
	return s
}

type Vulns []Vuln

func (vs Vulns) ToString() string {
	var s string
	for _, vuln := range vs {
		s += fmt.Sprintf("[ Find Vuln: %s ] ", vuln.ToString())
	}
	return s
}

func GetHeaderstr(resp *ReqRes) string {
	var headerstr = ""
	for k, v := range resp.Header {
		for _, i := range v {
			headerstr += fmt.Sprintf("%s: %s\r\n", k, i)
		}
	}
	return headerstr
}

func getFramework(result *IdentifyRes, resp *ReqRes, fingermap []*utils.Finger, matcher func(*IdentifyRes, *ReqRes, *utils.Finger)) {

	// 若默认端口未匹配到结果,则匹配全部
	for _, finger := range fingermap {
		matcher(result, resp, finger)
	}
	return
}

type Frameworks []Framework

func (fs Frameworks) ToString() string {
	framework_strs := make([]string, len(fs))
	for i, f := range fs {
		framework_strs[i] = f.ToString()
	}
	return strings.Join(framework_strs, "||")
}

func (fs Frameworks) GetTitles() []string {
	var titles []string
	//titles := []string{}
	for _, f := range fs {
		if !f.IsGuess {
			titles = append(titles, f.Name)
		}
	}
	return titles
}
