package Core

import (
	"Mule/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

var ResSlice []utils.PathDict

var FileLogger *zap.Logger
var ProBar *progressbar.ProgressBar
var CheckFlag int
var Format string

//初始化log

func InitLogger(logfile string, nolog, nobanner bool) {
	//defer utils.TimeCost()()
	//
	//fmt.Println("Start init logger")

	if !nolog {
		if !nobanner {
			fmt.Println(Mule)

			fmt.Println(Version)
		}

		writeSyncer, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0744)
		if err != nil {
			panic("create log_file failed")
		}

		//encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig := zapcore.EncoderConfig{
			MessageKey: "msg",
		}

		var encoder zapcore.Encoder

		if Format == "json" {
			encoder = zapcore.NewJSONEncoder(encoderConfig)
		} else {
			encoder = zapcore.NewConsoleEncoder(encoderConfig)
		}

		//Core := zapcore.NewCore(encoder,zapcore.NewMultiWriteSyncer(writeSyncer, zapcore.AddSync(os.Stdout)), zapcore.DebugLevel)
		core := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(writeSyncer), zapcore.DebugLevel)
		FileLogger = zap.New(core)
	}

}

func ResHandle(ctx context.Context, reschan chan *utils.PathDict) {
	for {
		select {
		case <-ctx.Done():
			return
		case res, ok := <-reschan:
			if !ok {
				return
			}
			ResSlice = append(ResSlice, *res)
		}
	}

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
			if err != nil {
				FileLogger.Error(dic + " open failed")
			}
			continue
		}
		if err1 := json.Unmarshal(bytes, &OldDict); err1 != nil {
			if err1 != nil {
				FileLogger.Error("Write json " + dic + " failed")
			}
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
			if FileLogger != nil {
				FileLogger.Error(errMarshall.Error())
			}
			return
		}

		if err := ioutil.WriteFile(NewDicPath, []byte(result), 0644); err != nil {
			if FileLogger != nil {
				FileLogger.Error("Write file " + NewDicPath + " error!")
			}
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
	var tmp int

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
			time.Sleep(200 * time.Millisecond)
			ProBar.Finish()
			return
		case _, ok := <-CountChan:
			if !ok {
				time.Sleep(200 * time.Millisecond)
				ProBar.Finish()
				return
			}

			tmp += 1
			if tmp%100 == 0 && tmp != 0 {
				//ProBar.Clear()
				ProBar.Add(100)
				tmp = 0
			}

		}
	}

}
