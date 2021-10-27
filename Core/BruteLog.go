package Core

import (
	"Mule/utils"
	"context"
	"encoding/json"
	"github.com/schollz/progressbar/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
)

var ResChan = make(chan utils.PathDict, 1000)

var ResSlice []utils.PathDict

var Logger *zap.Logger
var ProBar *progressbar.ProgressBar
var CheckFlag int

//初始化log

func InitLogger(logfile string) {
	//defer utils.TimeCost()()
	//
	//fmt.Println("Start init logger")

	writeSyncer, _ := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0744)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig = zapcore.EncoderConfig{
		MessageKey: "msg",
	}
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	//Core := zapcore.NewCore(encoder,zapcore.NewMultiWriteSyncer(writeSyncer, zapcore.AddSync(os.Stdout)), zapcore.DebugLevel)
	core := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(writeSyncer), zapcore.DebugLevel)
	Logger = zap.New(core)
}

func ResHandle(reschan chan utils.PathDict) []utils.PathDict {
	for res := range reschan {
		ResSlice = append(ResSlice, res)
	}

	return ResSlice

}

func OutputLinkFinder() {
	for fromurl, linklist := range SpiderUrlMap {
		parurl, _ := url.Parse(fromurl)
		op, _ := os.OpenFile(parurl.Host, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0744)
		op.WriteString(fromurl + "\n")
		for _, link := range linklist {
			op.WriteString("\t" + link + "\n")
		}
	}
	for fromurl, linklist := range SpiderJsMap {
		parurl, _ := url.Parse(fromurl)
		op, _ := os.OpenFile(parurl.Host, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0744)
		op.WriteString(fromurl + "\n")
		for _, link := range linklist {
			op.WriteString("\t" + link + "\n")
		}
	}
}

func UpdateDict(dicpath []string, dirroot string) {
	//defer utils.TimeCost()()
	//fmt.Println("start update dict")

	DictMap := make(map[string][]string)

	for _, res := range ResSlice {
		DictMap[res.Tag] = append(DictMap[res.Tag], res.Path)
	}

	for _, dic := range dicpath {
		var OldDict, NewDict []utils.PathInfo
		filename, ext := utils.GetNameSuffix(dic)

		NewDicPath := filepath.Join(dirroot, "Data", "DefDict", filename+".json")

		if ext == "" {
			dic = filepath.Join(dirroot, "Data", "DefDict", filename+".json")

		}

		bytes, err := ioutil.ReadFile(dic)

		if err != nil {
			Logger.Error(dic + " open failed")
			continue
		}
		if err1 := json.Unmarshal(bytes, &OldDict); err1 != nil {
			Logger.Error("Write json " + dic + " failed")
			continue
		}

		// 用来给hits加1的的地方

		for _, m := range OldDict {
			if StringInSlice(m.Path, DictMap[filename]) {
				m.Hits += 1
				NewDict = append(NewDict, m)
			} else {
				NewDict = append(NewDict, m)
			}
		}

		// 最后面4个空格，让json格式更美观
		//result, errMarshall := json.MarshalIndent(newJson, "", "    ")
		// 最后面4个空格，让json格式更美观
		result, errMarshall := utils.CustomMarshal(NewDict)

		if errMarshall != nil {
			Logger.Error(errMarshall.Error())
			return
		}

		if err := ioutil.WriteFile(NewDicPath, []byte(result), 0644); err != nil {
			Logger.Error("Write file " + NewDicPath + " error!")
			return
		}
	}
}

//  判断字符是否在字符列表中
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func IntInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func BruteProcessBar(ctx context.Context, PathCap int, Target string, CountChan chan struct{}) {
	// create and start new bar

	ProBar = progressbar.NewOptions(PathCap,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetDescription("[cyan] Now Processing [reset]"+Target),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=",
			SaucerPadding: " ",
			BarStart:      "|",
			BarEnd:        "|",
			SaucerHead:    "[blue]>",
		}))

	for {
		select {
		case <-ctx.Done():
			ProBar.Finish()
			return
		case _, ok := <-CountChan:
			if !ok {
				ProBar.Finish()
				return
			}

			ProBar.Add(1)

		}
	}

}
