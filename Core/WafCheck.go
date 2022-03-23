package Core

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
)

// 定期检测是否触发了安全设备被block了
func TriggerWaf(ctx context.Context, client *CustomClient, target string, wd *WildCard) (bool, error) {

	WafTest, err := client.RunRequest(ctx, target, Additional{
		Mod:   "default",
		Value: RandomPath,
	})

	if err != nil {
		return true, errors.New("Request_Error")
	}

	comres, err := CompareWildCard(wd, WafTest)

	if !comres {
		return false, nil
	}

	return true, err

}

func timeChecking(ctx context.Context, client *CustomClient, target string, wd *WildCard, callback, ck chan struct{}, ctxcancel context.CancelFunc) {

	checkFlag := 0
	cancelFlag := 0
	timeoutflag := 0
	checktime := 50
	immcheck := false
	for {

		select {
		case <-ctx.Done():
			return
		case _, ok := <-ck:
			if !ok {
				return
			}
			checkFlag += 1
			if checkFlag%checktime == 0 && checkFlag != 0 {
				immcheck = true
			}

		case _, ok := <-callback:
			if !ok {
				return
			}
			immcheck = true

		}
		if immcheck {
			immcheck = false
		ReCon:
			res, err := TriggerWaf(ctx, client, target, wd)
			//fmt.Println(len(repChan))
			if err != nil {
				//fmt.Printf("bad luck, you have been blocked %s, there is a waf or check your network\n", target)
				//ctxcancel()
				if err.Error() == "Request_Error" {
					timeoutflag += 1
					if timeoutflag >= 3 {
						FileLogger.Error("failed",
							zap.String("error_target", target),
							zap.Int("error_item", checkFlag))
						//暂时block报错输出
						waf := fmt.Sprintf("\nbad luck, you have been blocked %s, now item %v\n", target, checkFlag)
						fmt.Fprintln(Pgbar.Bypass(), waf)
						ctxcancel()
						return
					}
					goto ReCon
				} else {
					cancelFlag += 1
					goto ReCon
				}

			} else if res {
				cancelFlag += 1
				goto ReCon
			} else {
				// 在没有检测出问题时就直接增加1.5倍的下次check时间，降低消耗
				checktime = checktime + checktime/2
			}

			if cancelFlag >= Block {
				FileLogger.Error("failed",
					zap.String("error_target", target),
					zap.Int("error_item", checkFlag))
				//暂时block报错输出
				fmt.Fprintln(Pgbar.Bypass(), fmt.Sprintf("\nbad luck, you have been blocked %s, now item %v\n", target, checkFlag))
				ctxcancel()
				return
			}
		}
	}

}
