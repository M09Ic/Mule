package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var (
	src = rand.NewSource(time.Now().UnixNano())
)

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func TimeCost() func() {
	start := time.Now()
	return func() {
		tc := time.Since(start)
		fmt.Printf("time cost = %v\n", tc)
	}
}

type PathDict struct {
	PathInfo
	Tag string
}

type PathInfo struct {
	Path string `json:"path"`
	Hits int    `json:"hits"`
}

func RandStringBytesMaskImprSrcUnsafe(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

func HandleTarget(target string) (string, error) {
	if !strings.HasPrefix(target, "http") {
		// check to see if a port was specified
		re := regexp.MustCompile(`^[^/]+:(\d+)`)
		match := re.FindStringSubmatch(target)

		if len(match) < 2 {
			// no port, default to http on 80
			target = fmt.Sprintf("http://%s", target)
		} else {
			port, err2 := strconv.Atoi(match[1])
			if err2 != nil || (port != 80 && port != 443) {
				return target, fmt.Errorf("url scheme not specified")
			} else if port == 80 {
				target = fmt.Sprintf("http://%s", target)
			} else {
				target = fmt.Sprintf("https://%s", target)
			}
		}
	}
	//if !strings.HasSuffix(target,"/"){
	//	target = target + "/"
	//}

	if strings.HasSuffix(target, "/") {
		target = target[:len(target)-1]
	}

	return target, nil
}

func HandleLocation(location string) (string, error) {
	u, err := url.Parse(location)
	if err != nil {
		return location, err
	}

	path := u.Path

	// path???????????????????????????,?????????java?????????????????????path?????????;jsessionid?????????,??????url??????????????????
	if strings.Contains(u.Path, ";jsessionid") {
		index := strings.Index(u.Path, ";jsessionid")
		path = u.Path[:index]
	}

	handled := u.Host + path

	return handled, nil
}

func ReadDict(info []string, root string, rang string, noupdate bool) []PathDict {
	/*
		????????????????????????????????????????????????????????????
	*/

	//defer TimeCost()()
	if Nolog {
		fmt.Println("start read dict")
	}

	var allJson []PathDict
	var err error

	for _, dictpath := range info {
		var eachJson []PathInfo

		tagname, pathext := GetNameSuffix(dictpath)

		if pathext == "" {
			dictpath = filepath.Join(root, "Data", "DefDict", dictpath+".json")
		} else if pathext == ".txt" && !noupdate {
			dictpath, err = TextToJsonOfFile(dictpath, tagname, root)

			if err != nil {
				fmt.Printf("can't convert %s to json\n", dictpath)
				continue
			}

		}
		dictbytes, err := ioutil.ReadFile(dictpath)

		if !noupdate {
			if err != nil {
				println(dictpath + " open failed")
				//panic(dictPath + " open failed")
				continue
			}

			if err := json.Unmarshal(dictbytes, &eachJson); err != nil {
				println(" Unmarshal failed")
				continue
			}

			for _, y := range eachJson {
				mid := PathDict{
					PathInfo: y,
					Tag:      tagname,
				}
				allJson = append(allJson, mid)
			}
		} else {
			dict := string(dictbytes)
			dicts := strings.Split(dict, "\n")
			for _, p := range dicts {
				mid := PathDict{
					PathInfo: PathInfo{
						Path: strings.TrimSpace(p),
						Hits: 0,
					},
					Tag: "",
				}
				allJson = append(allJson, mid)
			}
		}

		if Nolog {
			fmt.Println("use dict: " + dictpath)
		}

	}

	// ?????????json????????????Hits????????????
	sort.Slice(allJson,
		func(i, j int) bool {
			return allJson[i].Hits > allJson[j].Hits
		})

	if rang == "0" {
		return allJson
	} else if !strings.Contains(rang, "-") {
		EndRang, err := strconv.Atoi(rang)
		if err != nil {
			panic("please check End")
		}
		return allJson[:EndRang]
	} else if strings.Contains(rang, "-") {
		RangList := strings.Split(rang, "-")
		Ben, err := strconv.Atoi(RangList[0])
		if err != nil {
			panic("please check End")
		}
		if RangList[1] == "" {
			return allJson[Ben:]
		}
		End, err := strconv.Atoi(RangList[1])
		if err != nil {
			panic("please check End")
		}
		return allJson[Ben:End]
	}

	return allJson

}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetNameSuffix(filename string) (string, string) {
	filenameWithSuffix := path.Base(filename)

	fileSuffix := path.Ext(filenameWithSuffix)

	filenameOnly := strings.TrimSuffix(filenameWithSuffix, fileSuffix)

	return filenameOnly, fileSuffix
}

func CustomMarshal(message interface{}) (string, error) {
	/*
		????????????????????????????????? "&"??????????????????
	*/

	bf := bytes.NewBuffer([]byte{})

	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	jsonEncoder.SetIndent("", "    ")

	if err := jsonEncoder.Encode(message); err != nil {
		return "", err
	}

	return bf.String(), nil
}

func GetDefaultList(def string) (DefList []string) {
	if def != "" {
		if strings.Contains(def, ",") {
			userslice := strings.Split(def, ",")
			for _, user := range userslice {
				DefList = append(DefList, user)
			}
		} else {
			DefList = append(DefList, def)
		}
	}
	return DefList
}

func ReadTarget(targetfile string) (TargetList []string, err error) {
	file, err := os.Open(targetfile)
	if err != nil {
		panic("please check your file")
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		user := strings.TrimSpace(scanner.Text())
		if user != "" {
			TargetList = append(TargetList, user)
		}
	}
	return TargetList, err
}

func FilterNewLines(s string) string {
	return regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(strings.TrimSpace(s), " ")
}

func GetExtType(rawUrl string) string {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return ""
	}
	return path.Ext(u.Path)
}
