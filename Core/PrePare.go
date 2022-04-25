package Core

import (
	"Mule/utils"
	"context"
	"fmt"
	"github.com/gosuri/uiprogress"
	"github.com/panjf2000/ants"
	"github.com/projectdiscovery/cdncheck"
	"net"
	"regexp"
	"strings"
	"sync"
)

type PreParePara struct {
	client    *CustomClient
	preclient *CustomClient
	target    string
	root      string
}

func FilterTarget(ctx context.Context, client, preclient *CustomClient, targets []string, root string) {

	Pgbar = uiprogress.New()
	Pgbar.Start()

	var Prep sync.WaitGroup
	PreParePool, _ := ants.NewPoolWithFunc(50, func(Para interface{}) {
		CuPara := Para.(PreParePara)
		ScanPrepare(ctx, &CuPara)
		defer Prep.Done()
	})

	for _, tar := range targets {
		Prep.Add(1)
		pp := PreParePara{
			preclient: client,
			client:    client,
			target:    tar,
			root:      root,
		}
		_ = PreParePool.Invoke(pp)

	}

	Prep.Wait()
	PreParePool.Release()
	close(Targetwd)
}

func ScanPrepare(ctx context.Context, para *PreParePara) {

	fmt.Println("Start to check " + para.target)

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

	resp, target, aliverr := CheckProto(ctx, para.target, para.preclient)

	if !aliverr {
		fmt.Fprintln(Pgbar.Bypass(), fmt.Sprintln("cannot connect to "+para.target))
		return
	}
	HandleredTarget = append(HandleredTarget, target)
	RandomPath = utils.RandStringBytesMaskImprSrcUnsafe(12)

	//wildcard, err := client.RunRequest(ctx, target+"/"+RandomPath)

	WdMap, err := GenWildCardMap(ctx, para.client, RandomPath, target, para.root)

	if err != nil {
		fmt.Fprintln(Pgbar.Bypass(), fmt.Sprintln("cannot connect to "+para.target))
		return
	}
	if resp != nil {
		WdMap["defaultcon"] = &WildCard{
			StatusCode: resp.StatusCode,
			Body:       resp.Body,
			Length:     resp.Length,
			Type:       0,
		}
	}

	var tw tarwp

	tw.target = para.target
	tw.wildmap = WdMap
	Targetwd <- tw

	return

}

func CheckProto(ctx context.Context, target string, client *CustomClient) (*ReqRes, string, bool) {
	temptarget, err := utils.HandleTarget(target)
	var res *ReqRes

	if err != nil {
		return nil, "", false
	}

	res, err = client.RunRequest(ctx, temptarget, Additional{
		Mod:   "default",
		Value: "",
	})

	if err != nil {
		if strings.HasPrefix(temptarget, "https://") {
			return nil, "", false
		} else if strings.HasPrefix(temptarget, "http://") {
			temptarget = strings.Replace(temptarget, "http", "https", 1)
			res, err = client.RunRequest(ctx, temptarget, Additional{
				Mod:   "default",
				Value: "",
			})
			if err != nil {
				return nil, "", false
			}
		}
	}

	if strings.HasPrefix(temptarget, "http://") {
		if (res.StatusCode >= 300 && res.StatusCode < 400) && strings.HasPrefix(res.Header.Get("Location"), "https") {
			temptarget = strings.Replace(temptarget, "http", "https", 1)
		} else if res.StatusCode == 400 {
			temptarget = strings.Replace(temptarget, "http", "https", 1)
			res, err = client.RunRequest(ctx, temptarget, Additional{
				Mod:   "default",
				Value: "",
			})
			if err != nil {
				return nil, "", false
			}
		}
	}

	return res, temptarget, true
}
