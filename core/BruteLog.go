package core

import (
	"Mule/utils"
	"encoding/json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"os"
	"path/filepath"
)


var ResChan = make(chan utils.PathDict, 10)

var ResSlice []utils.PathDict

//初始化log

func InitLogger(logfile string) {
	writeSyncer, _ := os.Create(logfile)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig = zapcore.EncoderConfig{
		MessageKey: "msg",
	}
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	//core := zapcore.NewCore(encoder,zapcore.NewMultiWriteSyncer(writeSyncer, zapcore.AddSync(os.Stdout)), zapcore.DebugLevel)
	core := zapcore.NewCore(encoder,zapcore.NewMultiWriteSyncer(writeSyncer), zapcore.DebugLevel)
	Logger = zap.New(core)
}


func ResHandle(reschan chan utils.PathDict) []utils.PathDict{
	for res := range reschan{
		ResSlice = append(ResSlice, res)
	}

	return ResSlice

}

func UpdateDict(dicpath []string,dirroot string){

	DictMap := make(map[string][]string)

	for _, res := range ResSlice {
		DictMap[res.Tag] = append(DictMap[res.Tag], res.Path)
	}

	for _ ,dic := range dicpath {
		var OldDict, NewDict []utils.PathDict
		filename, ext := utils.GetNameSuffix(dic)



		NewDicPath := filepath.Join(dirroot, "Data", filename+"_new"+".json")

		if ext == ""{
			dic = filepath.Join(dirroot, "Data", filename +".json")

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