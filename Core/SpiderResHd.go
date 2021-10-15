package Core

import (
	"Mule/utils"
	"strings"
)

var SpiderBlackWord = []string{"javascript:"}
var SpiderJsMap = map[string][]string{}
var SpiderUrlList []string

var SpiderChan = make(chan *SpiderRes, 1000)

type SpiderRes struct {
	Loc    string
	Path   string
	JsPath []string
}

func SpiderResHandle(sp chan *SpiderRes) {
	for Spres := range sp {
		switch Spres.Loc {
		case "orgin":
			if !ContainsInBk(Spres.Path, SpiderBlackWord) {
				SpiderUrlList = append(SpiderUrlList, Spres.Path)
			}
		case "js":
			Spres.JsPath = utils.RemoveDuplicateElement(Spres.JsPath)
			SpiderJsMap[Spres.Loc] = Spres.JsPath
		}
	}

}

func ContainsInBk(tester string, bk []string) bool {
	for _, black := range bk {
		if strings.Contains(tester, black) {
			return true
		}
	}
	return false
}
