package Core

import (
	"Mule/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/panjf2000/ants"
	"github.com/projectdiscovery/cdncheck"
	"net"
	"regexp"
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

type PoolPara struct {
	ctx      context.Context
	wordchan chan utils.PathDict
	custom   *CustomClient
	target   string
	wgs      *sync.WaitGroup
	wdmap    map[string]*WildCard
}

func ScanPrepare(ctx context.Context, client *CustomClient, target string, root string) (map[string]*WildCard, error) {

	//defer utils.TimeCost()()
	//fmt.Println("start scan prepare")
	var WdMap map[string]*WildCard

	_, err := client.RunRequest(ctx, target, Additional{
		Mod:   "default",
		Value: "",
	})

	if err != nil {
		return nil, fmt.Errorf("cann't connect to %s", target)
	}

	RandomPath = utils.RandStringBytesMaskImprSrcUnsafe(12)

	//wildcard, err := client.RunRequest(ctx, target+"/"+RandomPath)

	WdMap, err = GenWildCardMap(ctx, client, RandomPath, target, root)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return WdMap, nil

}

func ScanTask(ctx context.Context, Opts Options, client *CustomClient) error {

	taskroot, cancel := context.WithCancel(context.Background())
	defer cancel()

	var SpWgs sync.WaitGroup
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

	alljson := utils.ReadDict(Opts.Dictionary, Opts.DirRoot, Opts.Range, Opts.NoUpdate)

	utils.Configloader()
	for _, curtarget := range Opts.Target {
		CheckFlag = 0

		//t1 := time.Now()
		CurContext, CurCancel = context.WithCancel(taskroot)

		// ????????????????????????????????????????????????????????????????????????????????????
		wildcardmap, err := ScanPrepare(ctx, client, curtarget, Opts.DirRoot)

		if err != nil {
			fmt.Println(err)
			continue
		}

		//fmt.Println("Start to clean channel")

		select {
		case <-CheckChan:

		case <-Countchan:
		case <-RepChan:
		default:
		}

		// ??????cdn??????
		r, _ := regexp.Compile("((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})(\\.((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})){3}")
		if len(r.Find([]byte(curtarget))) != 0 {
			ipv4 := string(r.Find([]byte(curtarget)))
			client, err := cdncheck.NewWithCache()
			if err == nil {

				if found, err := client.Check(net.ParseIP(ipv4)); found && err == nil {
					fmt.Printf("%v is a part of cdn, so pass", ipv4)
					continue
				}
			}

		}

		//????????????????????????
		WordChan := MakeWordChan(alljson)
		//????????????????????????????????????
		// waf??????
		go TimingCheck(CurContext, client, curtarget, wildcardmap["default"], CheckChan, CurCancel)
		//?????????
		if Opts.Nolog {
			//go BruteProcessBar(CurContext, PathLength, curtarget, Countchan)
		}

		//  ???????????????
		ReqScanPool, _ := ants.NewPoolWithFunc(Opts.Thread, func(Para interface{}) {
			CuPara := Para.(PoolPara)
			AccessWork(&CuPara)
		})

		RepScanPool, _ := ants.NewPoolWithFunc(Opts.Thread, func(Para interface{}) {
			CuPara := Para.(ResponsePara)
			AccessResponseWork(&CuPara)
		})

		var ReqWgs, RepWgs sync.WaitGroup

		PrePara := PoolPara{
			ctx:      CurContext,
			wordchan: WordChan,
			custom:   client,
			target:   curtarget,
			wgs:      &ReqWgs,
			wdmap:    wildcardmap,
		}

		RespPre := ResponsePara{
			ctx:      CurContext,
			repchan:  RepChan,
			wgs:      &RepWgs,
			wdmap:    wildcardmap,
			mod:      Opts.Mod,
			jsfinder: Opts.JsFinder,
		}

		//??????????????????
		go ResHandle(ResChan)

		for i := 0; i < Opts.Thread; i++ {
			ReqWgs.Add(1)
			RepWgs.Add(1)
			_ = RepScanPool.Invoke(RespPre)
			_ = ReqScanPool.Invoke(PrePara)
		}

		//????????????
		ReqWgs.Wait()
		RepWgs.Wait()
		if Opts.JsFinder {
			fmt.Println("????????????????????????linkfinder????????????")
			time.Sleep(500 * time.Millisecond)
			SpWgs.Wait()
		}

		//elapsed := time.Since(t1)
		//fmt.Println("App elapsed: ", elapsed)

		// TODO ??????hits??????json
		//for _, i := range ResSlice{
		//	fmt.Println(i.Path)
		//}

		if Opts.JsFinder {
			OutputLinkFinder()
		}
		if !Opts.Nolog {
			JsonRes, _ := json.Marshal(ResSlice)
			fmt.Println(string(JsonRes))
		} else if !Opts.NoUpdate {
			UpdateDict(Opts.Dictionary, Opts.DirRoot)
		}
		CurCancel()
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

func AccessWork(WorkPara *PoolPara) {
	defer WorkPara.wgs.Done()
	//result,err := custom.RunRequest(ctx, Url)

	for {
		select {
		case <-WorkPara.ctx.Done():
			return

		case word, ok := <-WorkPara.wordchan:
			if !ok {
				return
			}

			if utils.Nolog {
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

			result, err := WorkPara.custom.RunRequest(WorkPara.ctx, WorkPara.target, add)

			if err != nil {
				// TODO ????????????
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

			RepChan <- &curresp

		}
	}

}
