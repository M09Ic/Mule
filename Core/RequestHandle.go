package Core

import (
	"Mule/utils"
	"context"
	"strings"
	"time"
)

func AccessWork(ctx context.Context, workPara *PoolPara) {
	defer workPara.wgs.Done()
	//result,err := custom.RunRequest(ctx, Url)
	for {
		select {
		case <-ctx.Done():
			return

		case word, ok := <-workPara.wordchan:
			if !ok {
				return
			}

			if !utils.Nolog {
				workPara.countchan <- struct{}{}
			}

			workPara.checkChan <- struct{}{}

			path := word.Path

			PreHandleWord := strings.TrimSpace(path)
			if strings.HasPrefix(PreHandleWord, "#") || len(PreHandleWord) == 0 {
				continue
			}

			if workPara.custom.Mod == "default" {
				if !strings.HasPrefix(PreHandleWord, "/") {
					PreHandleWord = "/" + PreHandleWord
				}
			}

			add := Additional{
				Mod:   workPara.custom.Mod,
				Value: PreHandleWord,
			}

			result, err := workPara.custom.RunRequest(ctx, workPara.target, add)

			if err != nil {
				// TODO 错误处理
				continue
			}

			curresp := Resp{
				resp: result,
				finpath: handledpath{
					target:        workPara.target,
					preHandleWord: PreHandleWord,
				},
				path: word,
			}

			select {
			case <-ctx.Done():
				return
			default:
				select {
				case workPara.repChan <- &curresp:
				case <-time.After(2 * time.Second):
					return
				}

			}

		}
	}

}

func CloseStopch(stop chan struct{}) {
	defer func() {
		if recover() != nil {
			// 返回值可以被修改
			// 在一个延时函数的调用中。
		}
	}()
	close(stop)
	return
}
