package core

import (
	"Mule/utils"
	"context"
	"fmt"
	"github.com/antlabs/strsim"
	"github.com/panjf2000/ants"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

var RandomPath string
var PathLength int
var Countchan = make(chan struct{}, 200)

type PoolPara struct {
	ctx      context.Context
	wordchan chan utils.PathDict
	custom   *CustomClient
	target   string
	wgs      *sync.WaitGroup
	wd       *WildCard
}

func ScanPrepare(ctx context.Context, client *CustomClient, target string) (*ReqRes, error) {

	_, err := client.RunRequest(ctx, target)

	if err != nil {
		return nil, fmt.Errorf("cann't connect to %s", target)
	}

	RandomPath := utils.RandStringBytesMaskImprSrcUnsafe(12)

	// TODO 暂时是不以/结尾所以在随机资源这里加了一个斜杠
	wildcard, err := client.RunRequest(ctx, target+"/"+RandomPath)

	if err != nil {
		return nil, err
	}

	return wildcard, nil

}

func ScanTask(ctx context.Context, Opts Options, client *CustomClient) error {

	t1 := time.Now()

	//读取字典返回管道
	WordChan := MakeWordChan(Opts.Dictionary, Opts.DirRoot)

	// 做访问前准备，判断是否可以连通，以及不存在路径的返回情况

	wildcard, err := ScanPrepare(ctx, client, Opts.Target)

	if err != nil {
		return err
	}

	// 完成对不存在页面的处理

	wd, err := HandleWildCard(wildcard)

	go BruteProcessBar(ctx, PathLength, Opts.Target, Countchan)

	//  开启线程池
	ScanPool, _ := ants.NewPoolWithFunc(Opts.Thread, func(Para interface{}) {
		CuPara := Para.(PoolPara)
		AccessWork(&CuPara)
	}, ants.WithExpiryDuration(2*time.Second))

	var wgs sync.WaitGroup

	PrePara := PoolPara{
		ctx:      ctx,
		wordchan: WordChan,
		custom:   client,
		target:   Opts.Target,
		wgs:      &wgs,
		wd:       wd,
	}

	go ResHandle(ResChan)

	for i := 0; i < Opts.Thread; i++ {
		wgs.Add(1)
		_ = ScanPool.Invoke(PrePara)
	}

	time.Sleep(500 * time.Millisecond)

	wgs.Wait()

	elapsed := time.Since(t1)
	fmt.Println("App elapsed: ", elapsed)

	// TODO 根据hits更新json
	//for _, i := range ResSlice{
	//	fmt.Println(i.Path)
	//}
	UpdateDict(Opts.Dictionary, Opts.DirRoot)

	return nil
}

//func MakeWordChan(DicPath string) chan string {
//	file, err := os.Open(DicPath)
//
//	WordChan := make(chan string)
//
//	if err != nil {
//		panic("please check your dictionary")
//	}
//
//	buf := bufio.NewReader(file)
//
//	go func() {
//		for {
//			line, _, err := buf.ReadLine()
//			if err == io.EOF {
//				break
//			}
//			WordChan <- string(line)
//
//		}
//		file.Close()
//		close(WordChan)
//	}()
//
//	return WordChan
//}

func MakeWordChan(DicSlice []string, DirRoot string) chan utils.PathDict {
	WordChan := make(chan utils.PathDict)

	alljson := utils.ReadDict(DicSlice, DirRoot)

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

func HandleWildCard(wildcard *ReqRes) (*WildCard, error) {

	if wildcard.StatusCode == 200 {
		wd := WildCard{
			StatusCode: wildcard.StatusCode,
			Body:       wildcard.Body,
			Length:     int64(len(wildcard.Body)),
			Type:       1,
		}
		return &wd, nil

	} else if wildcard.StatusCode > 300 && wildcard.StatusCode < 404 {
		wd := WildCard{
			StatusCode: wildcard.StatusCode,
			Location:   wildcard.Header.Get("Location"),
			Type:       2,
		}
		return &wd, nil
	} else {
		wd := WildCard{
			StatusCode: wildcard.StatusCode,
			Type:       3,
		}
		return &wd, nil
	}

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

			Countchan <- struct{}{}

			path := word.Path

			PreHandleWord := strings.TrimSpace(path)
			if strings.HasPrefix(PreHandleWord, "#") || len(PreHandleWord) == 0 {
				break
			}

			result, err := WorkPara.custom.RunRequest(WorkPara.ctx, WorkPara.target+PreHandleWord)

			if err != nil {
				// TODO 错误处理
				continue
			}

			// 和资源不存在页面进行比较
			comres, err := CompareWildCard(WorkPara.wd, result)

			if comres {
				ProBar.Clear()
				fmt.Printf("Path:%s \t Code:%d\n", WorkPara.target+PreHandleWord, result.StatusCode)
				word.Hits += 1
				Logger.Info("Success",
					zap.String("Path", WorkPara.target+PreHandleWord),
					zap.Int("Code", result.StatusCode))
				ResChan <- word
			}

		}
	}

}

func CompareWildCard(wd *WildCard, result *ReqRes) (bool, error) {
	switch wd.Type {

	// 类型1即资源不存在页面状态码为200
	case 1:
		if result.StatusCode == 200 {
			comres, err := Compare200(&wd.Body, &result.Body)
			return comres, err
		} else if (result.StatusCode > 300 && result.StatusCode < 404) || result.StatusCode == 503 {
			return true, nil
		}
	// 类型2即资源不存在页面状态码为30x
	case 2:
		if result.StatusCode == 200 {
			return true, nil
		} else if result.StatusCode == wd.StatusCode {
			comres, err := Compare30x(wd.Location, result.Header.Get("Location"))
			return comres, err
		} else if (result.StatusCode > 300 && result.StatusCode < 404) || result.StatusCode == 503 {
			return true, nil
		}
		// 类型3 即资源不存在页面状态码404或者奇奇怪怪
	case 3:
		// TODO 存在nginx的类似301跳转后的404页面
		if wd.StatusCode != result.StatusCode {
			if result.StatusCode == 200 || (result.StatusCode > 300 && result.StatusCode < 404) || result.StatusCode == 503 {
				return true, nil
			}
		}

	}

	return false, nil

}

func Compare30x(WdLoc string, Res string) (bool, error) {
	// 改为url对比 会出现由于参数问题导致的location对比不合理
	ratio := 0.98

	HandleWd, err := utils.HandleLocation(WdLoc)

	if err != nil {
		HandleWd = WdLoc
	}

	HandleRes, err := utils.HandleLocation(Res)

	if err != nil {
		HandleRes = Res
	}

	ComRatio := strsim.Compare(HandleWd, HandleRes)

	if ratio > ComRatio {
		return true, nil
	}

	return false, nil

}

func Compare200(WdBody *[]byte, ResBody *[]byte) (bool, error) {
	ratio := 0.98

	ComRatio := strsim.Compare(string(*WdBody), string(*ResBody))

	if ratio > ComRatio {
		return true, nil
	}

	return false, nil
}
