package Core

import (
	"Mule/utils"
	"context"
	"fmt"
	"github.com/fatih/color"
	"go.uber.org/zap"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type ResponsePara struct {
	repchan  chan *Resp
	resChan  chan *utils.PathDict
	StopCh   chan struct{}
	wgs      *sync.WaitGroup
	wdmap    map[string]*WildCard
	cachemap *sync.Map
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

func AccessResponseWork(ctx context.Context, WorkPara *ResponsePara) {
	defer WorkPara.wgs.Done()
	//result,err := custom.RunRequest(ctx, Url)

	//TODO channel超时
	for {
		select {
		case <-ctx.Done():
			return

		case <-WorkPara.StopCh:
			return
		case resp, ok := <-WorkPara.repchan:
			if !ok {
				CloseStopch(WorkPara.StopCh)
				return
			}
			// 和资源不存在页面进行比较

			if resp.resp.Header.Get("Location") != "" {
				resp.Hash = utils.Md5Hash(utils.BytesCombine(resp.resp.Body, []byte(resp.resp.Header.Get("Location"))))
			} else {
				resp.Hash = utils.Md5Hash(resp.resp.Body)
			}

			comres, _ := CustomCompare(WorkPara.wdmap, resp.finpath.preHandleWord, resp, WorkPara.cachemap)
			//comres, err := CompareWildCard(WorkPara.wdmap["default"], result)

			if comres {
				switch WorkPara.mod {
				case "default":
					finpath := resp.finpath.target + resp.finpath.preHandleWord
					resp.Alivepath = finpath
					temp := resp.Hash
					fingeriden := identifyResp(resp.resp)
					resp.IdentifyRes = fingeriden
					resp.path.Hits += 1
					fingeriden.Hash = temp

					if FileLogger != nil {
						if Format == "json" {
							FileLogger.Info("success",
								zap.String("path", finpath),
								zap.Int("code", resp.resp.StatusCode),
								zap.Int64("length", resp.resp.Length),
								zap.String("mmh3", fingeriden.Mmh3),
								zap.String("md5", fingeriden.Hash),
								zap.String("sim3", fingeriden.SimHash),
								zap.String("frameworks", fingeriden.Frameworks.ToString()))
						} else {
							FileLogger.Info(fmt.Sprintf("Path: %s\t%v\t%v\t[Framework:%s]\n", finpath, resp.resp.StatusCode, resp.resp.Length, fingeriden.Frameworks.ToString()))
						}
					}
					if !utils.Noconsole {
						//err := ProBar.Clear()
						//if err != nil {
						//	return
						//}
						blue := color.New(color.FgBlue).SprintFunc()
						cy := color.New(color.FgCyan).SprintFunc()
						red := color.New(color.FgHiMagenta).SprintFunc()
						io.WriteString(os.Stdout, fmt.Sprintf("\r%s\r", strings.Repeat(" ", 40)))

						fmt.Printf("Path: %s\t%s\t%s\t[Framework:%s]\n", blue(finpath), cy(resp.resp.StatusCode), red(resp.resp.Length), cy(fingeriden.Frameworks.ToString()))
					}
					select {
					case <-ctx.Done():
						return
					default:
						select {
						case WorkPara.resChan <- &resp.path:
							continue
						case <-time.After(5 * time.Second):
							return
						}

					}

				case "host":
					fingeriden := identifyResp(resp.resp)
					if !utils.Noconsole {
						blue := color.New(color.FgBlue).SprintFunc()
						cy := color.New(color.FgCyan).SprintFunc()
						red := color.New(color.FgHiMagenta).SprintFunc()
						io.WriteString(os.Stdout, fmt.Sprintf("\r%s\r", strings.Repeat(" ", 40)))

						fmt.Printf("IP: %s \tHost: %s \t%s\t%s\n", cy(resp.finpath.target), blue(resp.finpath.preHandleWord), cy(resp.resp.StatusCode), red(resp.resp.Length))
					}
					resp.path.Hits += 1
					if FileLogger != nil {

						//err := ProBar.Clear()
						//if err != nil {
						//	return
						//}
						if Format == "json" {
							FileLogger.Info("success",
								zap.String("ip", resp.finpath.target),
								zap.String("path", resp.finpath.preHandleWord),
								zap.Int("code", resp.resp.StatusCode),
								zap.Int64("length", resp.resp.Length),
								zap.String("mmh3", fingeriden.Mmh3),
								zap.String("md5", fingeriden.Hash),
								zap.String("sim3", fingeriden.SimHash),
								zap.String("frameworks", fingeriden.Frameworks.ToString()))
						} else {
							FileLogger.Info(fmt.Sprintf("IP: %v \tHost: %v\t%v\t%v", resp.finpath.target, resp.finpath.preHandleWord, resp.resp.StatusCode, resp.resp.Length))
						}

					}
					select {
					case <-ctx.Done():
						return
					default:

						select {
						case WorkPara.resChan <- &resp.path:
						case <-time.After(5 * time.Second):
							return
						}
					}
				}

				if WorkPara.jsfinder && WorkPara.mod == "default" {
					SpiderWaitChan <- resp.finpath.target + resp.finpath.preHandleWord
				}

			}

		}
	}

}
