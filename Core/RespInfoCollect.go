package Core

import (
	"Mule/utils"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

func identifyResp(resp *ReqRes) IdentifyRes {
	iden := IdentifyRes{}
	iden.Title = getTitle(string(resp.Body))
	iden.Title = EncodeTitle(iden.Title)
	if len(resp.Body) != 0 {
		iden.Mmh3 = utils.Mmh3Hash32(resp.Body)
		iden.SimHash = utils.Simhash(resp.Body)
	}

	getFramework(&iden, resp, utils.Httpfingers, httpFingerMatch)

	return iden
}

func EncodeTitle(s string) string {
	//if len(s) >= 13 {
	//	s = s[:13]
	//}
	s = strings.TrimSpace(s)
	s = fmt.Sprintf("%q", s)
	s = strings.Trim(s, "\"")
	return s
}

func getTitle(content string) string {
	if content == "" {
		return ""
	}
	title, ok := utils.CompileMatch(utils.CommonCompiled["title"], content)
	if ok {
		return title
	} else if len(content) > 13 {
		return content[0:13]
	} else {
		return content
	}
}

func httpFingerMatch(result *IdentifyRes, resp *ReqRes, finger *utils.Finger) {

	content := string(resp.Body)

	// 漏洞匹配优先
	for _, reg := range utils.Compiled[finger.Name+"_vuln"] {
		res, ok := utils.CompileMatch(reg, content)
		if ok {
			handlerMatchedResult(result, finger, res)
			result.AddFramework(Framework{Name: finger.Vuln})
			return
		}
	}
	// html匹配
	for _, body := range finger.Regexps.Body {
		if strings.Contains(content, body) {
			result.AddFramework(Framework{Name: finger.Name})
			return
		}
	}

	// 正则匹配
	for _, reg := range utils.Compiled[finger.Name] {
		res, ok := utils.CompileMatch(reg, content)
		if ok {
			handlerMatchedResult(result, finger, res)
			return
		}
	}
	// http头匹配
	for _, header := range finger.Regexps.Header {
		var headerstr string
		if resp == nil {
			headerstr = strings.ToLower(strings.Split(content, "\r\n\r\n")[0])
		} else {
			headerstr = strings.ToLower(GetHeaderstr(resp))
		}

		if strings.Contains(headerstr, strings.ToLower(header)) {
			result.AddFramework(Framework{Name: finger.Name})
			return
		}
	}

	// MD5 匹配
	for _, md5s := range finger.Regexps.MD5 {
		m := md5.Sum([]byte(content))
		if md5s == hex.EncodeToString(m[:]) {
			result.AddFramework(Framework{Name: finger.Name})
			return
		}
	}
}

func handlerMatchedResult(result *IdentifyRes, finger *utils.Finger, res string) {
	result.AddFramework(Framework{Name: finger.Name, Version: res})
}
