package utils

import (
	"bufio"
	"bytes"
	"compress/flate"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/go-dedup/simhash"
	"github.com/twmb/murmur3"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var Nolog bool

func UnFlate(input []byte) []byte {
	rdata := bytes.NewReader(input)
	r := flate.NewReaderDict(rdata, []byte(flatedict))
	s, _ := ioutil.ReadAll(r)
	return s
}

func Decode(input string) []byte {
	b := Base64Decode(input)
	return UnFlate(b)
}

var flatedict = `,":`

func Flate(input []byte) []byte {
	var bf = bytes.NewBuffer([]byte{})
	var flater, _ = flate.NewWriterDict(bf, flate.BestCompression, []byte(flatedict))
	defer flater.Close()
	if _, err := flater.Write(input); err != nil {
		println(err.Error())
		return []byte{}
	}
	if err := flater.Flush(); err != nil {
		println(err.Error())
		return []byte{}
	}
	return bf.Bytes()
}

func Encode(input []byte) string {
	s := Flate(input)
	return Base64Encode(s)
}

func Base64Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	return data
}

func standBase64(braw []byte) []byte {
	bckd := base64.StdEncoding.EncodeToString(braw)
	var buffer bytes.Buffer
	for i := 0; i < len(bckd); i++ {
		ch := bckd[i]
		buffer.WriteByte(ch)
		if (i+1)%76 == 0 {
			buffer.WriteByte('\n')
		}
	}
	buffer.WriteByte('\n')
	return buffer.Bytes()
}

func Base64Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func TextToJsonOfFile(fileName string, fn string, root string) (string, error) {

	var newSlice []PathInfo

	infos, err := ReadLines(fileName)
	if err != nil {
		println(err.Error())
		return "", err
	}

	for _, v := range RemoveDuplicateElement(infos) {
		newTag := PathInfo{
			Path: v,
			Hits: 0,
		}
		newSlice = append(newSlice, newTag)

	}

	info, _ := CustomMarshal(newSlice)
	NewFilename := filepath.Join(root, "Data", "DefDict", fn+".json")
	err = ioutil.WriteFile(NewFilename, []byte(info), 0644)
	if err != nil {
		fmt.Println("Please check ")
	}

	return NewFilename, nil
}

// golang读取文件并且返回列表
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// 字符串列表去重函数
func RemoveDuplicateElement(infos []string) []string {
	result := make([]string, 0, len(infos))
	temp := map[string]struct{}{}
	for _, item := range infos {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func SliceContains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func CompileRegexp(s string) regexp.Regexp {
	reg, err := regexp.Compile(s)
	if err != nil {
		fmt.Println("[-] regexp string error: " + s)
		os.Exit(0)
	}
	return *reg
}

func CompileMatch(reg regexp.Regexp, s string) (string, bool) {
	res := reg.FindStringSubmatch(s)
	if len(res) == 1 {
		return "", true
	} else if len(res) == 2 {
		return strings.TrimSpace(res[1]), true
	}
	return "", false
}

func Md5Hash(raw []byte) string {
	m := md5.Sum(raw)
	return hex.EncodeToString(m[:])
}

func Mmh3Hash32(raw []byte) string {
	var h32 = murmur3.New32()
	_, _ = h32.Write(standBase64(raw))
	return fmt.Sprintf("%d", h32.Sum32())
}

func Simhash(raw []byte) string {

	sh := simhash.NewSimhash()
	return fmt.Sprintf("%x", sh.GetSimhash(sh.NewWordFeatureSet(raw)))
}

func MaptoString(m map[string]interface{}) string {
	if m == nil || len(m) == 0 {
		return ""
	}
	var s string
	for k, v := range m {
		s += fmt.Sprintf(" %s:%s ", k, ToString(v))
	}
	return s
}

func ToStringMap(i interface{}) map[string]string {
	var m = map[string]string{}

	switch v := i.(type) {
	case map[interface{}]interface{}:
		for k, val := range v {
			m[ToString(k)] = ToString(val)
		}
		return m
	case map[string]interface{}:
		for k, val := range v {
			m[k] = ToString(val)
		}
		return m
	default:
		return nil
	}
}

func ToString(data interface{}) string {
	switch s := data.(type) {
	case nil:
		return ""
	case string:
		return s
	case bool:
		return strconv.FormatBool(s)
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(s), 'f', -1, 32)
	case int:
		return strconv.Itoa(s)
	case int64:
		return strconv.FormatInt(s, 10)
	case int32:
		return strconv.Itoa(int(s))
	case int16:
		return strconv.FormatInt(int64(s), 10)
	case int8:
		return strconv.FormatInt(int64(s), 10)
	case uint:
		return strconv.FormatUint(uint64(s), 10)
	case uint64:
		return strconv.FormatUint(s, 10)
	case uint32:
		return strconv.FormatUint(uint64(s), 10)
	case uint16:
		return strconv.FormatUint(uint64(s), 10)
	case uint8:
		return strconv.FormatUint(uint64(s), 10)
	case []byte:
		return string(s)
	case fmt.Stringer:
		return s.String()
	case error:
		return s.Error()
	default:
		return fmt.Sprintf("%v", data)
	}
}
