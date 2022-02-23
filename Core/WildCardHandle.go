package Core

import (
	"Mule/utils"
	"context"
	"github.com/antlabs/strsim"
	"path/filepath"
	"strings"
	"sync"
)

var RandomPath string
var BlackList []int

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
		} else if (!IntInSlice(result.StatusCode, BlackList) && result.StatusCode != 400) && ((result.StatusCode > 300 && result.StatusCode < 404) || result.StatusCode == 503 || result.StatusCode == 500) {
			return true, nil
		}
	// 类型2即资源不存在页面状态码为30x
	case 2:
		if result.StatusCode == 200 {
			return true, nil
		} else if result.StatusCode == wd.StatusCode {
			comres, err := Compare30x(wd.Location, strings.Split(result.Header.Get("Location"), ";")[0])
			return comres, err
		} else if (!IntInSlice(result.StatusCode, BlackList) && result.StatusCode != 400) && ((result.StatusCode > 300 && result.StatusCode < 404) || result.StatusCode == 503) {
			return true, nil
		}
		// 类型3 即资源不存在页面状态码404或者奇奇怪怪
	case 3:
		if wd.StatusCode != result.StatusCode {
			if (!IntInSlice(result.StatusCode, BlackList) && result.StatusCode != 400) && (result.StatusCode == 200 || (result.StatusCode > 300 && result.StatusCode < 404) || result.StatusCode == 503 || result.StatusCode == 500) {
				return true, nil
			}
		}

	}

	return false, nil

}

func CustomCompare(wdmap map[string]*WildCard, path string, result *Resp, cachemap *sync.Map) (bool, error) {

	// 修改的false的原因在于如果是重复的页面出现可以直接说明目标页面不存在了，或是重复的页面，一次来解决二级目录问题
	// TODO 是否存在一些爆破需求是目标界面为500这样的通用报错

	if !(result.resp.StatusCode >= 403 && result.resp.StatusCode <= 503) {
		if _, ok := cachemap.Load(result.Hash); ok {
			//fmt.Println("use cache")
			return false, nil
		}
	}

	for key := range wdmap {
		if key == "default" {
			continue
		} else {
			if strings.Contains(path, key) {
				res, err := CompareWildCard(wdmap[key], result.resp)
				if result.resp.StatusCode >= 403 && result.resp.StatusCode <= 503 {
					return res, err
				}
				cachemap.Store(result.Hash, res)
				return res, err
			}
		}
	}

	key := "default"
	res, err := CompareWildCard(wdmap[key], result.resp)
	cachemap.Store(result.Hash, res)
	return res, err

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
			Location:   strings.Split(wildcard.Header.Get("Location"), ";")[0],
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

func GenWildCardMap(ctx context.Context, client *CustomClient, random string, target string, proroot string) (map[string]*WildCard, error) {
	var Testpath string
	var wd *WildCard
	resmap := make(map[string]*WildCard)

	switch client.Mod {
	case "default":
		wdlist, err := GetExPathList(proroot)

		wdlist = append(wdlist, "")

		if err != nil {
			return nil, err
		}

		for _, ex := range wdlist {
			if ex == "" {
				Testpath = "/" + RandomPath
				wd, err = GenWd(ctx, client, target, Testpath)
				if err != nil {
					return nil, err
				}
				resmap["default"] = wd

			} else if !strings.Contains(ex, "$$") {
				wd_tmp := WildCard{
					StatusCode: 403,
					Type:       3,
				}
				resmap[ex[1:]] = &wd_tmp
			} else {
				Testpath = strings.Replace(ex, "$$", random, 1)
				wd, err = GenWd(ctx, client, target, Testpath)
				if err != nil {
					//TODO 增加一个访问的黑名单，如果ban掉的目录就直接不访问
					continue
					//return nil, fmt.Errorf("When you test %s, there is something error\n", ex)
				}
				in := strings.Index(ex, "$$")
				key := ex[in+2:]
				resmap[key] = wd
			}
		}
	case "host":
		Testpath = "/"
		wd, err := GenWd(ctx, client, target, Testpath)
		if err != nil {
			return nil, err
		}
		resmap["default"] = wd
	}

	return resmap, nil

}

func GetExPathList(root string) ([]string, error) {
	// TODO 将加入参数dirroot,这里为了测试方便使用了绝对路径
	expath := filepath.Join(root, "Data", "SpecialList", "exwildcard.txt")

	ex, err := utils.ReadLines(expath)
	if err != nil {
		println(err.Error())
		return nil, err
	}

	return ex, nil
}

func GenWd(ctx context.Context, client *CustomClient, target string, Tpath string) (*WildCard, error) {
	wildcard, err := client.RunRequest(ctx, target, Additional{
		Mod:   "default",
		Value: Tpath,
	})

	if err != nil {
		return nil, err
	}

	wd, err := HandleWildCard(wildcard)

	return wd, nil
}
