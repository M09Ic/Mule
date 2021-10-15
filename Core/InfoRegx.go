package Core

import (
	"Mule/utils"
	"net/url"
	"regexp"
	"strings"
)

var AWSS3 = regexp.MustCompile(`(?i)[a-z0-9.-]+\.s3\.amazonaws\.com|[a-z0-9.-]+\.s3-[a-z0-9-]\.amazonaws\.com|[a-z0-9.-]+\.s3-website[.-](eu|ap|us|ca|sa|cn)|//s3\.amazonaws\.com/[a-z0-9._-]+|//s3-[a-z0-9-]+\.amazonaws\.com/[a-z0-9._-]+`)

var linkFinderRegex = regexp.MustCompile(`(?:"|')(((?:[a-zA-Z]{1,10}://|//)[^"'/]{1,}\.[a-zA-Z]{2,}[^"']{0,})|((?:/|\.\./|\./)[^"'><,;| *()(%%$^/\\\[\]][^"'><,;|()]{1,})|([a-zA-Z0-9_\-/]{1,}/[a-zA-Z0-9_\-/]{1,}\.(?:[a-zA-Z]{1,4}|action)(?:[\?|#][^"|']{0,}|))|([a-zA-Z0-9_\-/]{1,}/[a-zA-Z0-9_\-/]{3,}(?:[\?|#][^"|']{0,}|))|([a-zA-Z0-9_\-]{1,}\.(?:php|asp|aspx|jsp|json|action|html|js|txt|xml)(?:[\?|#][^"|']{0,}|)))(?:"|')`)

func LinkFinder(source string) ([]string, error) {
	var links []string
	// source = strings.ToLower(source)
	if len(source) > 1000000 {
		source = strings.ReplaceAll(source, ";", ";\r\n")
		source = strings.ReplaceAll(source, ",", ",\r\n")
	}
	source = DecodeChars(source)

	match := linkFinderRegex.FindAllStringSubmatch(source, -1)
	for _, m := range match {
		matchGroup1 := utils.FilterNewLines(m[1])
		if matchGroup1 == "" {
			continue
		}
		links = append(links, matchGroup1)
	}
	links = utils.RemoveDuplicateElement(links)
	return links, nil
}

func GetAWSS3(source string) []string {
	var aws []string
	for _, match := range AWSS3.FindAllStringSubmatch(source, -1) {
		aws = append(aws, DecodeChars(match[0]))
	}
	return aws
}

//转义处理
func DecodeChars(s string) string {
	source, err := url.QueryUnescape(s)
	if err == nil {
		s = source
	}

	// In case json encoded chars
	replacer := strings.NewReplacer(
		`\u002f`, "/",
		`\u0026`, "&",
	)
	s = replacer.Replace(s)
	return s
}
