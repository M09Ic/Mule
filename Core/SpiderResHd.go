package Core

import (
	"Mule/utils"
	"strings"
)

var SpiderBlackWord = []string{"javascript:", "jquery"}
var SpiderJsMap = map[string][]string{}
var SpiderUrlMap = map[string][]string{}

var SpiderChan = make(chan *SpiderRes, 1000)

type SpiderRes struct {
	Orgin  string
	Loc    string
	Path   string
	JsPath []string
}

func SpiderResHandle(sp chan *SpiderRes) {
	for Spres := range sp {
		switch Spres.Loc {
		case "orgin":
			if _, ok := SpiderUrlMap[Spres.Orgin]; ok {
				if !ContainsInBk(Spres.Path, SpiderBlackWord) {
					SpiderUrlMap[Spres.Orgin] = append(SpiderUrlMap[Spres.Orgin], Spres.Path)
				}
			} else {
				SpiderUrlMap[Spres.Orgin] = append([]string{}, Spres.Path)
			}

		case "js":
			Spres.JsPath = utils.RemoveDuplicateElement(Spres.JsPath)
			SpiderJsMap[Spres.Orgin] = Spres.JsPath
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
