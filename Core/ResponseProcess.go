package Core

import (
	"Mule/utils"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

type ResponsePara struct {
	ctx      context.Context
	repchan  chan *Resp
	wgs      *sync.WaitGroup
	wdmap    map[string]*WildCard
	mod      string
	jsfinder bool
}

type Resp struct {
	resp      *ReqRes
	finpath   handledpath
	path      utils.PathDict
	Alivepath string `json:"Path"`
	IdentifyRes
}

type handledpath struct {
	target        string
	preHandleWord string
}

func AccessResponseWork(WorkPara *ResponsePara) {
	defer WorkPara.wgs.Done()
	//result,err := custom.RunRequest(ctx, Url)

	//TODO channel超时
	for {
		select {
		case <-WorkPara.ctx.Done():
			return

		case <-time.After(3 * time.Second):
			return
		case resp, ok := <-WorkPara.repchan:
			if !ok {
				return
			}
			// 和资源不存在页面进行比较
			comres, _ := CustomCompare(WorkPara.wdmap, resp.finpath.preHandleWord, resp.resp)
			//comres, err := CompareWildCard(WorkPara.wdmap["default"], result)

			if comres {
				switch WorkPara.mod {
				case "default":
					finpath := resp.finpath.target + resp.finpath.preHandleWord
					resp.Alivepath = finpath
					fingeriden := identifyResp(resp.resp)
					resp.IdentifyRes = fingeriden
					resp.path.Hits += 1

					if FileLogger != nil {
						//ProBar.Clear()
						//blue := color.New(color.FgBlue).SprintFunc()
						//cy := color.New(color.FgCyan).SprintFunc()
						//red := color.New(color.FgHiMagenta).SprintFunc()
						//fmt.Printf("Path: %s\tCode: %s\tLength: %s\t[Framework:%s]\n", blue(finpath), cy(resp.resp.StatusCode), red(resp.resp.Length), cy(fingeriden.Frameworks.ToString()))
						FileLogger.Info("success",
							zap.String("path", finpath),
							zap.Int("code", resp.resp.StatusCode),
							zap.Int64("length", resp.resp.Length),
							zap.String("title", fingeriden.Title),
							zap.String("mmh3", fingeriden.Mmh3),
							zap.String("md5", fingeriden.Hash),
							zap.String("simhash", fingeriden.SimHash),
							zap.String("frameworks", fingeriden.Frameworks.ToString()))
					}
					ResChan <- resp
				case "host":
					//ProBar.Clear()
					//blue := color.New(color.FgBlue).SprintFunc()
					//cy := color.New(color.FgCyan).SprintFunc()
					//red := color.New(color.FgHiMagenta).SprintFunc()
					//fmt.Printf("IP: %s \tPath: %s \t Code:%s \t Length:%s\n", cy(resp.finpath.target), blue(resp.finpath.preHandleWord), cy(resp.resp.StatusCode), red(resp.resp.Length))
					resp.path.Hits += 1
					FileLogger.Info("Success",
						zap.String("ip", resp.finpath.target),
						zap.String("host", resp.finpath.preHandleWord),
						zap.Int("Code", resp.resp.StatusCode),
						zap.Int64("Length", resp.resp.Length))
					ResChan <- resp
				}

				if WorkPara.jsfinder && WorkPara.mod == "default" {
					SpiderWaitChan <- resp.finpath.target + resp.finpath.preHandleWord
				}

			}

		}
	}

}

func identifyResp(resp *ReqRes) IdentifyRes {
	iden := IdentifyRes{}
	iden.Title = getTitle(string(resp.Body))
	iden.Title = EncodeTitle(iden.Title)
	if len(resp.Body) != 0 {
		iden.Hash = utils.Md5Hash(resp.Body)
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

	//} else if finger.Regexps.Cookie != nil {
	//	for _, cookie := range finger.Regexps.Cookie {
	//		if resp == nil {
	//			if strings.Contains(content, cookie) {
	//				result.Frameworks = finger.Name
	//				return
	//			}
	//		} else if cookies[cookie] != "" {
	//			result.Frameworks = finger.Name
	//			return
	//		}
	//	}/
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
