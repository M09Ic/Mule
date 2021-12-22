package utils

import "regexp"

var CommonCompiled map[string]regexp.Regexp

func Configloader() {
	Compiled = make(map[string][]regexp.Regexp)
	Mmh3fingers, Md5fingers = LoadHashFinger()
	Httpfingers = LoadFingers("http")
	CommonCompiled = map[string]regexp.Regexp{
		"title":     CompileRegexp("(?Uis)<title>(.*)</title>"),
		"server":    CompileRegexp("(?i)Server: ([\x20-\x7e]+)"),
		"xpb":       CompileRegexp("(?i)X-Powered-By: ([\x20-\x7e]+)"),
		"sessionid": CompileRegexp("(?i) (.*SESS.*?ID)"),
	}
}
