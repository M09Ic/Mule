package Core

import (
	"Mule/utils"
	"fmt"
)

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
		var last string
		switch Spres.Loc {
		case "orgin":
			if last == "" {
				last = Spres.Orgin
				SpiderUrlMap[Spres.Orgin] = append(SpiderUrlMap[Spres.Orgin], Spres.Path)
			} else if last == Spres.Orgin {
				SpiderUrlMap[Spres.Orgin] = append(SpiderUrlMap[Spres.Orgin], Spres.Path)
			} else {
				JsLogger.Info(fmt.Sprintf("%s", last))
				for _, link := range SpiderUrlMap[last] {
					JsLogger.Info(fmt.Sprintf("\t%s", link))
				}
				delete(SpiderUrlMap, last)
				last = Spres.Orgin
				SpiderUrlMap[Spres.Orgin] = append(SpiderUrlMap[Spres.Orgin], Spres.Path)
			}

		case "js":
			Spres.JsPath = utils.RemoveDuplicateElement(Spres.JsPath)
			JsLogger.Info(fmt.Sprintf("%s", Spres.Orgin))
			for _, link := range Spres.JsPath {
				JsLogger.Info(fmt.Sprintf("\t%s", link))
			}
		}
	}

}

//func OutputLinkFinder() {
//	for fromurl, linklist := range SpiderUrlMap {
//		parurl, _ := url.Parse(fromurl)
//		op, _ := os.OpenFile("./log/"+parurl.Host, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0744)
//		op.WriteString(fromurl + "\n")
//		for _, link := range linklist {
//			op.WriteString("\t" + link + "\n")
//		}
//	}
//	for fromurl, linklist := range SpiderJsMap {
//		parurl, _ := url.Parse(fromurl)
//		op, _ := os.OpenFile("./log/"+parurl.Host, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0744)
//		op.WriteString(fromurl + "\n")
//		for _, link := range linklist {
//			op.WriteString("\t" + link + "\n")
//		}
//	}
//}
