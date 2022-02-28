package Core

import (
	"Mule/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/panjf2000/ants"
	"strconv"
	"strings"
	"sync"
	"time"
)

var PathLength int
var SpiderWaitChan = make(chan string, 100)
var HandleredTarget []string
var Block int
var AllWildMap sync.Map

type PoolPara struct {
	wordchan  chan utils.PathDict
	StopCh    chan struct{}
	countchan chan struct{}
	checkChan chan struct{}
	repChan   chan *Resp
	custom    *CustomClient
	target    string
	wgs       *sync.WaitGroup
	wdmap     map[string]*WildCard
}

type tarwp struct {
	wildmap map[string]*WildCard
	target  string
}

func ScanTask(ctx context.Context, Opts Options, client, preclient *CustomClient) error {

	//f, err := os.OpenFile("cpu.prof", os.O_RDWR|os.O_CREATE, 0644)
	//if err != nil {
	//	panic(err)
	//}
	//pprof.StartCPUProfile(f)

	taskroot, cancel := context.WithCancel(ctx)
	defer cancel()

	//var SpWgs sync.WaitGroup

	////js探测
	//if Opts.JsFinder {
	//	go SpiderResHandle(SpiderChan)
	//	SpiderScanPool, _ := ants.NewPoolWithFunc(10, func(Para interface{}) {
	//		defer SpWgs.Done()
	//		CuPara := Para.(string)
	//		newcolly := NewCollyClient(&Opts)
	//		newcolly.Start(CuPara)
	//		newcolly.NormalCollector.Wait()
	//		newcolly.LinkFinderCollector.Wait()
	//	})
	//
	//	go func(Spchan chan string) {
	//		for i := range Spchan {
	//			SpWgs.Add(1)
	//			_ = SpiderScanPool.Invoke(i)
	//		}
	//	}(SpiderWaitChan)
	//
	//}

	Opts.Target = GetRange(Opts.TargetRange, Opts.Target)

	FilterTarget(taskroot, client, preclient, Opts.Target, Opts.DirRoot)

	alljson := utils.ReadDict(Opts.Dictionary, Opts.DirRoot, Opts.Range, Opts.NoUpdate, Opts.AutoDict)

	var wg sync.WaitGroup
	TaskPool, _ := ants.NewPoolWithFunc(Opts.PoolSize, func(Para interface{}) {
		defer wg.Done()
		CuPara := Para.(WorkPara)
		StartProcess(taskroot, &CuPara)
	})
	utils.Configloader()
	var targetwd []tarwp

	for _, curtarget := range HandleredTarget {
		var wildcardmap map[string]*WildCard

		var tw tarwp

		if wd, ok := AllWildMap.Load(curtarget); !ok {
			continue
		} else {
			wildcardmap = wd.(map[string]*WildCard)
			tw.target = curtarget
			tw.wildmap = wildcardmap
			targetwd = append(targetwd, tw)
		}
	}

	if len(targetwd) > 1 {
		Opts.Thread = Opts.Thread / 2
	}

	for _, curtargetwd := range targetwd {

		fmt.Println("Start brute " + curtargetwd.target)

		wp := WorkPara{
			alljson: &alljson,
			Opts:    Opts,
			client:  client,
			target:  curtargetwd.target,
			wdmap:   curtargetwd.wildmap,
		}

		wg.Add(1)
		_ = TaskPool.Invoke(wp)

		//t1 := time.Now()

		// 做访问前准备，判断是否可以连通，以及不存在路径的返回情况

	}
	wg.Wait()
	return nil
}

type WorkPara struct {
	alljson *[]utils.PathDict
	Opts    Options
	client  *CustomClient
	target  string
	wdmap   map[string]*WildCard
}

func StartProcess(ctx context.Context, wp *WorkPara) {

	repChan := make(chan *Resp, 1000)
	resChan := make(chan *utils.PathDict, 1000)
	checkChan := make(chan struct{}, 1000)
	CurContext, CurCancel := context.WithCancel(ctx)

	//读取字典返回管道
	WordChan := MakeWordChan(*wp.alljson)
	//检测成功后初始化各类插件
	// waf检测
	go timeChecking(CurContext, wp.client, wp.target, wp.wdmap["default"], checkChan, CurCancel)
	//进度条
	countchan := make(chan struct{}, 1000)
	if !wp.Opts.Nolog {
		go BruteProcessBar(CurContext, PathLength, wp.target, countchan)
	}

	//  开启线程池
	ReqScanPool, _ := ants.NewPoolWithFunc(wp.Opts.Thread, func(Para interface{}) {
		CuPara := Para.(PoolPara)
		AccessWork(CurContext, &CuPara)
	})
	RepScanPool, _ := ants.NewPoolWithFunc(wp.Opts.Thread, func(Para interface{}) {
		CuPara := Para.(ResponsePara)
		AccessResponseWork(CurContext, &CuPara)
	})

	var ReqWgs, RepWgs sync.WaitGroup
	var StopCh_R = make(chan struct{})
	PrePara := PoolPara{
		wordchan:  WordChan,
		repChan:   repChan,
		checkChan: checkChan,
		countchan: countchan,
		custom:    wp.client,
		target:    wp.target,
		wgs:       &ReqWgs,
		StopCh:    StopCh_R,
		wdmap:     wp.wdmap,
	}

	cm := sync.Map{}

	RespPre := ResponsePara{
		repchan:  repChan,
		resChan:  resChan,
		wgs:      &RepWgs,
		wdmap:    wp.wdmap,
		cachemap: &cm,
		mod:      wp.Opts.Mod,
		StopCh:   StopCh_R,
		jsfinder: wp.Opts.JsFinder,
	}

	//开启结果协程
	go ResHandle(CurContext, resChan)

	for i := 0; i < wp.Opts.Thread; i++ {
		ReqWgs.Add(1)
		RepWgs.Add(1)
		_ = RepScanPool.Invoke(RespPre)
		_ = ReqScanPool.Invoke(PrePara)
	}

	ReqWgs.Wait()
	go ReqScanPool.Release()
	for {
		if len(repChan) == 0 {
			close(repChan)
			break
		}
	}

	StopCh := make(chan struct{})
	go func() {
		RepWgs.Wait()
		close(StopCh)
	}()

	select {
	case <-StopCh:
		fmt.Printf("%v finsihed\n", wp.target)

	case <-time.After(time.Duration(wp.Opts.Timeout+2) * time.Second):
		fmt.Printf("%v break of time\n", wp.target)

	}

	RepScanPool.Release()

	//if Opts.JsFinder {
	//	fmt.Println("扫描结束，请等待linkfinder运行结束")
	//	time.Sleep(500 * time.Millisecond)
	//	SpWgs.Wait()
	//}
	//
	//if Opts.JsFinder {
	//	OutputLinkFinder()
	//}
	if wp.Opts.Nolog {
		JsonRes, _ := json.Marshal(ResSlice)
		fmt.Println(string(JsonRes))
	}
	if !wp.Opts.NoUpdate {
		UpdateDict(wp.Opts.Dictionary, wp.Opts.DirRoot)
	}
	//pprof.StopCPUProfile()
	//f.Close()
	select {
	case <-CurContext.Done():
		return
	default:
		CurCancel()
		return
	}
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

func GetRange(rang string, allJson []string) []string {
	if rang == "0" {
		return allJson
	} else if !strings.Contains(rang, "-") {
		End, err := strconv.Atoi(rang)
		if err != nil {
			panic("please check End")
		}
		if End >= len(allJson) {
			fmt.Println("out of range,it's set to the end")
			End = len(allJson)
		}
		return allJson[:End]
	} else if strings.Contains(rang, "-") {
		RangList := strings.Split(rang, "-")
		Ben, err := strconv.Atoi(RangList[0])
		if err != nil {
			panic("please check End")
		}
		if RangList[1] == "" {
			return allJson[Ben:]
		}
		End, err := strconv.Atoi(RangList[1])
		if err != nil {
			panic("please check End")
		}
		if End >= len(allJson) {
			fmt.Println("out of range,it's set to the end")
			End = len(allJson)
		}
		return allJson[Ben:End]
	}
	return allJson
}
