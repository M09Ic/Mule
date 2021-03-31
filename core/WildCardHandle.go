package core

import (
	"Mule/utils"
	"context"
	"github.com/antlabs/strsim"
)

var RandomPath string

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

func GenWildCard(ctx context.Context, client *CustomClient, target string) {

}
