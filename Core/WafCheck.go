package Core

import (
	"context"
	"fmt"
)

var CancelFlag int

// 定期检测是否触发了安全设备被block了
func TriggerWaf(ctx context.Context, client *CustomClient, target string, wd *WildCard) (bool, error) {

	WafTest, err := client.RunRequest(ctx, target, Additional{
		Mod:   "default",
		Value: RandomPath,
	})

	if err != nil {
		return true, fmt.Errorf("bad luck, you have been blocked %s, there is a waf or check your network", target)
	}

	comres, err := CompareWildCard(wd, WafTest)

	if !comres {
		return false, nil
	}

	return true, err
	// 一段时间后访问相同url,如果状态码不一样则触发waf,(true为触发waf)
	//if wd.StatusCode != WafTest.StatusCode {
	//	return true, nil
	//}
	//
	//c3, err1 := Compare30x(wd.Location, WafTest.Header.Get("Location"))
	//c2, _ := Compare200(&wd.Body, &WafTest.Body)
	//// 对比location失败则只判断body情况,如果一样,就返回false
	//if err1 != nil {
	//	return !c2, nil
	//}
	//
	////如果都对比没问题,就需要跳转一致且body一致就返回false
	//if c3 && c2 {
	//	return false, nil
	//}
	//
	//return true, nil

}

func TimingCheck(ctx context.Context, client *CustomClient, target string, wd *WildCard, ck chan struct{}, ctxcancel context.CancelFunc) {
	CheckFlag = 0
	CancelFlag = 0
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-ck:
			if !ok {
				return
			}
			CheckFlag += 1
			if CheckFlag%100 == 0 && CheckFlag != 0 {
				res, err := TriggerWaf(ctx, client, target, wd)
				//fmt.Println(len(RepChan))
				if err != nil {
					//fmt.Printf("bad luck, you have been blocked %s, there is a waf or check your network\n", target)
					//ctxcancel()
					CancelFlag += 1
				} else if res {
					//fmt.Printf("bad luck, you have been blocked %s, there is a waf or check your network\n", target)
					//ctxcancel()
					CancelFlag += 1
				}

				if CancelFlag > Block {
					//FileLogger.Info("failed",
					//	zap.String("Path", ""),
					//	zap.Int("Code", 0),
					//	zap.Int64("Length", 0))
					//暂时block报错输出
					fmt.Printf("\nbad luck, you have been blocked %s, there is a waf or check your network\n", target)
					ctxcancel()
				}
			}

		}
	}
}
