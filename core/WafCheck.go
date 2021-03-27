package core

import (
	"context"
	"fmt"
)

//中途判断waf应该加个参数是否开启
func TriggerWaf(ctx context.Context, client *CustomClient, target string, wd *WildCard) (bool, error) {

	WafTest, err := client.RunRequest(ctx, RandomPath)

	if err != nil {
		return true, fmt.Errorf("bad luck, you have been blocked %s, there is a waf or check your network", target)
	}

	// 一段时间后访问相同url,如果状态码不一样则触发waf,(true为触发waf)
	if wd.StatusCode != WafTest.StatusCode {
		return true, nil
	}

	c3, err1 := Compare30x(wd.Location, WafTest.Header.Get("Location"))
	c2, _ := Compare200(&wd.Body, &WafTest.Body)
	// 对比location失败则只判断body情况,如果一样,就返回false
	if err1 != nil {
		return !c2, nil
	}

	//如果都对比没问题,就需要跳转一致且body一致就返回false
	if c3 && c2 {
		return false, nil
	}

	return true, nil

}
