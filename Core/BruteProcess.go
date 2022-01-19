package Core

import (
	"Mule/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/panjf2000/ants"
	"strings"
	"sync"
	"time"
)

var PathLength int
var Countchan = make(chan struct{}, 10000)

var CurCancel context.CancelFunc
var CurContext context.Context
var CheckChan = make(chan struct{}, 10000)
var SpiderWaitChan = make(chan string, 100)
var RepChan = make(chan *Resp, 1000)
var Block int
var AllWildMap sync.Map

type PoolPara struct {
	wordchan chan utils.PathDict
	custom   *CustomClient
	target   string
	wgs      *sync.WaitGroup
	wdmap    map[string]*WildCard
}

func ScanTask(ctx context.Context, Opts Options, client *CustomClient) error {

	//f, err := os.OpenFile("cpu.prof", os.O_RDWR|os.O_CREATE, 0644)
	//if err != nil {
	//	panic(err)
	//}
	//pprof.StartCPUProfile(f)

	taskroot, cancel := context.WithCancel(ctx)
	defer cancel()

	var SpWgs sync.WaitGroup

	//js探测
	if Opts.JsFinder {
		go SpiderResHandle(SpiderChan)
		SpiderScanPool, _ := ants.NewPoolWithFunc(10, func(Para interface{}) {
			defer SpWgs.Done()
			CuPara := Para.(string)
			newcolly := NewCollyClient(&Opts)
			newcolly.Start(CuPara)
			newcolly.NormalCollector.Wait()
			newcolly.LinkFinderCollector.Wait()
		})

		go func(Spchan chan string) {
			for i := range Spchan {
				SpWgs.Add(1)
				_ = SpiderScanPool.Invoke(i)
			}
		}(SpiderWaitChan)

	}

	FilterTarget(taskroot, client, Opts.Target, Opts.DirRoot)

	alljson := utils.ReadDict(Opts.Dictionary, Opts.DirRoot, Opts.Range, Opts.NoUpdate)

	utils.Configloader()
	for _, curtarget := range Opts.Target {

		var wildcardmap map[string]*WildCard

		if wd, ok := AllWildMap.Load(curtarget); !ok {
			fmt.Println("cannot connect to " + curtarget)
			continue
		} else {
			fmt.Println("Start brute " + curtarget)
			wildcardmap = wd.(map[string]*WildCard)
		}

		CheckFlag = 0

		//t1 := time.Now()
		CurContext, CurCancel = context.WithCancel(taskroot)

		// 做访问前准备，判断是否可以连通，以及不存在路径的返回情况

		select {
		case <-CheckChan:

		case <-Countchan:
		case <-RepChan:

		case <-ResChan:
		default:
		}

		//读取字典返回管道
		WordChan := MakeWordChan(alljson)
		//检测成功后初始化各类插件
		// waf检测
		go TimingCheck(CurContext, client, curtarget, wildcardmap["default"], CheckChan, CurCancel)
		//进度条
		if !Opts.Nolog {
			go BruteProcessBar(CurContext, PathLength, curtarget, Countchan)
		}

		//  开启线程池
		ReqScanPool, _ := ants.NewPoolWithFunc(Opts.Thread, func(Para interface{}) {
			CuPara := Para.(PoolPara)
			AccessWork(CurContext, &CuPara)
		})

		RepScanPool, _ := ants.NewPoolWithFunc(Opts.Thread, func(Para interface{}) {
			CuPara := Para.(ResponsePara)
			AccessResponseWork(CurContext, &CuPara)
		})

		var ReqWgs, RepWgs sync.WaitGroup

		PrePara := PoolPara{
			wordchan: WordChan,
			custom:   client,
			target:   curtarget,
			wgs:      &ReqWgs,
			wdmap:    wildcardmap,
		}

		cm := sync.Map{}

		RespPre := ResponsePara{
			repchan:  RepChan,
			wgs:      &RepWgs,
			wdmap:    wildcardmap,
			cachemap: &cm,
			mod:      Opts.Mod,
			jsfinder: Opts.JsFinder,
		}

		//开启结果协程
		go ResHandle(CurContext, ResChan)

		for i := 0; i < Opts.Thread; i++ {
			ReqWgs.Add(1)
			RepWgs.Add(1)
			_ = RepScanPool.Invoke(RespPre)
			_ = ReqScanPool.Invoke(PrePara)
		}

		ReqWgs.Wait()
		ReqScanPool.Release()

		StopCh := make(chan struct{})

		go func() {
			RepWgs.Wait()
			close(StopCh)
		}()

		select {
		case <-StopCh:
			fmt.Println("break of close")
			break
		case <-time.After(time.Duration(Opts.Timeout+2) * time.Second):
			fmt.Println("break of time")
			break
		}

		RepScanPool.Release()
		if Opts.JsFinder {
			fmt.Println("扫描结束，请等待linkfinder运行结束")
			time.Sleep(500 * time.Millisecond)
			SpWgs.Wait()
		}

		if Opts.JsFinder {
			OutputLinkFinder()
		}
		if Opts.Nolog {
			JsonRes, _ := json.Marshal(ResSlice)
			fmt.Println(string(JsonRes))
		}
		if !Opts.NoUpdate {
			UpdateDict(Opts.Dictionary, Opts.DirRoot)
		}
		//pprof.StopCPUProfile()
		//f.Close()
		select {
		case <-CurContext.Done():
			continue
		default:
			CurCancel()
		}
	}

	return nil
}

func MakeWordChan(alljson []utils.PathDict) chan utils.PathDict {
	WordChan := make(chan utils.PathDict)

	if len(alljson) == 0 {
		panic("please check your dict")
	}

	PathLength = len(alljson)

	go func() {
		for _, info := range alljson {
			WordChan <- info
		}

		close(WordChan)
	}()

	return WordChan
}

func AccessWork(ctx context.Context, WorkPara *PoolPara) {
	defer WorkPara.wgs.Done()
	//result,err := custom.RunRequest(ctx, Url)

	for {
		select {
		case <-ctx.Done():
			return

		case word, ok := <-WorkPara.wordchan:
			if !ok {
				return
			}

			if !utils.Nolog {
				Countchan <- struct{}{}
			}

			CheckChan <- struct{}{}

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
				case RepChan <- &curresp:
				case <-time.After(2 * time.Second):
					return
				}

			}

		}
	}

}
