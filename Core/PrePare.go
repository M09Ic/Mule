package Core

import (
	"Mule/utils"
	"context"
	"fmt"
	"github.com/panjf2000/ants"
	"github.com/projectdiscovery/cdncheck"
	"net"
	"regexp"
	"sync"
)

type PreParePara struct {
	client *CustomClient
	target string
	root   string
}

func FilterTarget(ctx context.Context, client *CustomClient, targets []string, root string) {

	var Prep sync.WaitGroup
	PreParePool, _ := ants.NewPoolWithFunc(100, func(Para interface{}) {
		CuPara := Para.(PreParePara)
		ScanPrepare(ctx, &CuPara)
		defer Prep.Done()
	})

	for _, tar := range targets {
		Prep.Add(1)
		pp := PreParePara{

			client: client,
			target: tar,
			root:   root,
		}
		_ = PreParePool.Invoke(pp)
		//if err != nil {
		//	fmt.Println(err)
		//	continue
		//}

	}

	Prep.Wait()
	PreParePool.Release()

}

func ScanPrepare(ctx context.Context, para *PreParePara) {

	//defer utils.TimeCost()()
	//fmt.Println("start scan prepare")
	//var WdMap map[string]*WildCard

	fmt.Println("Start to check " + para.target)
	_, err := para.client.RunRequest(ctx, para.target, Additional{
		Mod:   "default",
		Value: "",
	})

	if err != nil {
		//fmt.Println(err)
		return
	}

	// 加入cdn检测
	r, _ := regexp.Compile("((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})(\\.((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})){3}")
	if len(r.Find([]byte(para.target))) != 0 {
		ipv4 := string(r.Find([]byte(para.target)))
		client, err := cdncheck.NewWithCache()
		if err == nil {

			if found, err := client.Check(net.ParseIP(ipv4)); found && err == nil {
				fmt.Printf("%v is a part of cdn, so pass\n", ipv4)
				return
			}
		}

	}

	RandomPath = utils.RandStringBytesMaskImprSrcUnsafe(12)

	//wildcard, err := client.RunRequest(ctx, target+"/"+RandomPath)

	WdMap, err := GenWildCardMap(ctx, para.client, RandomPath, para.target, para.root)

	if err != nil {
		//fmt.Println(err)
		return
	}

	AllWildMap.Store(para.target, WdMap)

	return

}
