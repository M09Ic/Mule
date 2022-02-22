package Core

import (
	"context"
	"strings"
	"time"
)

func AccessWork(ctx context.Context, WorkPara *PoolPara) {
	defer WorkPara.wgs.Done()
	//result,err := custom.RunRequest(ctx, Url)

	for {
		select {
		case <-ctx.Done():
			return

		case word, ok := <-WorkPara.wordchan:
			if !ok {
				CloseStopch(WorkPara.StopCh)
				return
			}

			//if !utils.Nolog {
			//	WorkPara.Countchan <- struct{}{}
			//}

			WorkPara.CheckChan <- struct{}{}

			path := word.Path

			PreHandleWord := strings.TrimSpace(path)
			if strings.HasPrefix(PreHandleWord, "#") || len(PreHandleWord) == 0 {
				continue
			}

			if !strings.HasPrefix(PreHandleWord, "/") {
				PreHandleWord = "/" + PreHandleWord
			}

			add := Additional{
				Mod:   WorkPara.custom.Mod,
				Value: PreHandleWord,
			}

			result, err := WorkPara.custom.RunRequest(ctx, WorkPara.target, add)

			if err != nil {
				// TODO 错误处理
				continue
			}

			curresp := Resp{
				resp: result,
				finpath: handledpath{
					target:        WorkPara.target,
					preHandleWord: PreHandleWord,
				},
				path: word,
			}

			//RepChan <- &curresp
			select {
			case <-ctx.Done():
				return
			default:
				select {
				case WorkPara.RepChan <- &curresp:
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
