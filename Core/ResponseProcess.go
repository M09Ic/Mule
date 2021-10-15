package Core

import (
	"Mule/utils"
	"context"
	"fmt"
	"github.com/fatih/color"
	"go.uber.org/zap"
	"sync"
)

type ResponsePara struct {
	ctx     context.Context
	repchan chan *Resp
	wgs     *sync.WaitGroup
	wdmap   map[string]*WildCard
	mod     string
}

type Resp struct {
	resp    *ReqRes
	finpath handledpath
	path    utils.PathDict
}

type handledpath struct {
	target        string
	PreHandleWord string
}

func AccessResponseWork(WorkPara *ResponsePara) {
	//defer WorkPara.wgs.Done()
	//result,err := custom.RunRequest(ctx, Url)

	//TODO channel超时
	for {
		select {
		case <-WorkPara.ctx.Done():
			return

		case resp, ok := <-WorkPara.repchan:
			if !ok {
				return
			}
			// 和资源不存在页面进行比较
			comres, _ := CustomCompare(WorkPara.wdmap, resp.finpath.PreHandleWord, resp.resp)
			//comres, err := CompareWildCard(WorkPara.wdmap["default"], result)

			if comres {
				switch WorkPara.mod {
				case "default":
					finpath := resp.finpath.target + resp.finpath.PreHandleWord
					ProBar.Clear()
					blue := color.New(color.FgBlue).SprintFunc()
					cy := color.New(color.FgCyan).SprintFunc()
					red := color.New(color.FgHiMagenta).SprintFunc()
					fmt.Printf("Path: %s \t Code:%s \t Length:%s\n", blue(finpath), cy(resp.resp.StatusCode), red(resp.resp.Length))
					resp.path.Hits += 1
					Logger.Info("Success",
						zap.String("Path", finpath),
						zap.Int("Code", resp.resp.StatusCode),
						zap.Int64("Length", resp.resp.Length))
					ResChan <- resp.path
				case "host":
					ProBar.Clear()
					blue := color.New(color.FgBlue).SprintFunc()
					cy := color.New(color.FgCyan).SprintFunc()
					red := color.New(color.FgHiMagenta).SprintFunc()
					fmt.Printf("IP: %s \tPath: %s \t Code:%s \t Length:%s\n", cy(resp.finpath.target), blue(resp.finpath.PreHandleWord), cy(resp.resp.StatusCode), red(resp.resp.Length))
					resp.path.Hits += 1
					Logger.Info("Success",
						zap.String("ip", resp.finpath.target),
						zap.String("host", resp.finpath.PreHandleWord),
						zap.Int("Code", resp.resp.StatusCode),
						zap.Int64("Length", resp.resp.Length))
					ResChan <- resp.path
				}

			}

		}
	}

}
