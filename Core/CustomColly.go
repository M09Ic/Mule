package Core

import (
	"Mule/utils"
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var DefaultDepth = 1
var CollyCache *sync.Map

type Crawler struct {
	// js的发现器
	LinkFinderCollector *colly.Collector
	Transport           *http.Transport
	Timeout             int
	// 基础的页面发现
	NormalCollector *colly.Collector
	site            *url.URL
}

func NewCollyClient(Opt *Options) *Crawler {

	newcolly := colly.NewCollector(
		//TODO 记得把sync开回来，异步加速
		colly.Async(true),
		colly.MaxDepth(DefaultDepth),
		colly.IgnoreRobotsTxt(),
	)

	if len(Opt.Headers) != 0 {
		for _, header := range Opt.Headers {
			newcolly.OnRequest(func(r *colly.Request) {
				r.Headers.Set(header.Name, header.Value)
			})
		}
	}

	extensions.RandomUserAgent(newcolly)

	err := newcolly.Limit(&colly.LimitRule{
		DomainGlob: "*",
		//colly的线程，限制直接写死为10
		Parallelism: 10,
		Delay:       time.Duration(Opt.Timeout) * time.Second,
		RandomDelay: time.Duration(Opt.Timeout) * time.Second,
	})

	if err != nil {
		panic("Failed to set Limit Rule")
	}

	return &Crawler{
		NormalCollector: newcolly,
		Transport:       Opt.Transport,
		Timeout:         Opt.Timeout,
	}
}

func (crawler *Crawler) Start(target string) {
	HdTarget, err := url.Parse(target)
	if err != nil {
		panic("please check your url")
	}
	spiderclient := &http.Client{
		Transport: crawler.Transport,
		Timeout:   time.Second * time.Duration(crawler.Timeout),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 最多跳转十次，以防出现设计失误导致的无限跳转
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			nextLocation := req.Response.Header.Get("Location")
			// Allow in redirect from http to https or in same hostname
			// We just check contain hostname or not because we set URLFilter in main collector so if
			// the URL is https://otherdomain.com/?url=maindomain.com, it will reject it
			if strings.Contains(nextLocation, HdTarget.Hostname()) {

				return nil
			}
			return http.ErrUseLastResponse
		},
	}
	crawler.site = HdTarget
	crawler.NormalCollector.SetClient(spiderclient)
	crawler.init()

	err = crawler.NormalCollector.Visit(crawler.site.String())

	if err != nil {
		panic("Failed to start ")
	}
}

func (crawler *Crawler) init() {
	// 初始化各类返回监听器
	//初始化js finder
	crawler.LinkFinderCollector = crawler.NormalCollector.Clone()
	crawler.LinkFinderCollector.OnResponse(func(response *colly.Response) {
		if response.StatusCode == 404 || response.StatusCode == 403 || response.StatusCode == 429 || response.StatusCode < 100 {
			return
		}

		respStr := string(response.Body)

		// Verify which link is working
		u := response.Request.URL.String()

		paths, err := LinkFinder(respStr)
		if err != nil {
			fmt.Println("something error in jsfinder")
			return
		}

		currentPathURL, err := url.Parse(u)
		currentPathURLerr := false
		if err != nil {
			currentPathURLerr = true
		}

		res := SpiderRes{
			Orgin:  response.Request.URL.String(),
			Loc:    "js",
			Path:   "",
			JsPath: paths,
		}
		SpiderChan <- &res

		// 输出js匹配结果
		fmt.Println(response.Request.URL.String())
		for _, relPath := range paths {
			// JS Regex Result
			//outputFormat = fmt.Sprintf("- %s", relPath)

			rebuildURL := ""
			if !currentPathURLerr {
				rebuildURL = FixUrl(currentPathURL, relPath)
			} else {
				rebuildURL = FixUrl(crawler.site, relPath)
			}

			if rebuildURL == "" || !LinkBlackList(rebuildURL) {
				continue
			}

			// Try to request JS path
			// Try to generate URLs with main site
			fileExt := utils.GetExtType(rebuildURL)
			if fileExt == ".js" || fileExt == ".xml" || fileExt == ".json" || fileExt == ".map" {
				crawler.feedLinkfinder(rebuildURL)
			}

			// Try to generate URLs with the site where Javascript file host in (must be in main or sub domain)
			//urlWithJSHostIn := FixUrl(crawler.site, relPath)
			//if urlWithJSHostIn != "" {
			//	fileExt := utils.GetExtType(urlWithJSHostIn)
			//	if fileExt == ".js" || fileExt == ".xml" || fileExt == ".json" || fileExt == ".map" {
			//		crawler.feedLinkfinder(urlWithJSHostIn)
			//	} else {
			//		_ = crawler.LinkFinderCollector.Visit(urlWithJSHostIn) //not print care for lost link
			//	}
			//}

		}
	})

	//初始化常规监听器
	// Handle js files，监听html中是否存在src，把所有的js提取出来再访问，进一步提取
	crawler.NormalCollector.OnHTML("[src]", func(e *colly.HTMLElement) {
		jsFileUrl := e.Request.AbsoluteURL(e.Attr("src"))
		jsFileUrl = FixUrl(crawler.site, jsFileUrl)
		if jsFileUrl == "" || !LinkBlackList(jsFileUrl) {
			return
		}

		fileExt := utils.GetExtType(jsFileUrl)
		if fileExt == ".js" || fileExt == ".xml" || fileExt == ".json" {
			crawler.feedLinkfinder(jsFileUrl)
		}
	})

	//监听normal
	crawler.NormalCollector.OnHTML("[href]", func(e *colly.HTMLElement) {
		urlString := e.Request.AbsoluteURL(e.Attr("href"))
		urlString = FixUrl(crawler.site, urlString)
		if urlString == "" || !LinkBlackList(urlString) {
			return
		}

		res := SpiderRes{
			Loc:    "orgin",
			Orgin:  crawler.site.String(),
			Path:   urlString,
			JsPath: nil,
		}
		SpiderChan <- &res
	})
	// 表单处理
	crawler.NormalCollector.OnHTML("form[action]", func(e *colly.HTMLElement) {
		formUrl := e.Request.URL.String()
		res := SpiderRes{
			Loc:    "orgin",
			Orgin:  crawler.site.String(),
			Path:   formUrl,
			JsPath: nil,
		}
		SpiderChan <- &res
	})

}

func (crawler *Crawler) feedLinkfinder(jsFileUrl string) {

	// If JS file is minimal format. Try to find original format

	if _, ok := CollyCache.Load(jsFileUrl); ok {
		return
	}
	CollyCache.Store(jsFileUrl, "1")

	if strings.Contains(jsFileUrl, ".min.js") {
		originalJS := strings.ReplaceAll(jsFileUrl, ".min.js", ".js")
		_ = crawler.LinkFinderCollector.Visit(originalJS)
	}

	// Send Javascript to Link Finder Collector

	_ = crawler.LinkFinderCollector.Visit(jsFileUrl)

}

func FixUrl(mainSite *url.URL, nextLoc string) string {
	nextLocUrl, err := url.Parse(nextLoc)
	if err != nil {
		return ""
	}
	return mainSite.ResolveReference(nextLocUrl).String()
}

//func (crawler *Crawler) findAWSS3(resp string) {
//	aws := GetAWSS3(resp)
//	for _, e := range aws {
//		if !crawler.awsSet.Duplicate(e) {
//			outputFormat := fmt.Sprintf("[aws-s3] - %s", e)
//			if crawler.JsonOutput {
//				sout := SpiderOutput{
//					Input:      crawler.Input,
//					Source:     "body",
//					OutputType: "aws",
//					Output:     e,
//				}
//				if data, err := jsoniter.MarshalToString(sout); err == nil {
//					outputFormat = data
//				}
//			}
//			fmt.Println(outputFormat)
//			if crawler.Output != nil {
//				crawler.Output.WriteToFile(outputFormat)
//			}
//		}
//	}
//}
