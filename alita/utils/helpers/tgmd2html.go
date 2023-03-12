package helpers

import (
	"fmt"
	"regexp"
	"strings"
)

// ReverseHTML2MD function to convert html formatted raw string to markdown to get noformat string
func ReverseHTML2MD(text string) string {
	Html2MdMap := map[string]string{
		"i":    "_%s_",
		"u":    "__%s__",
		"b":    "*%s*",
		"s":    "~%s~",
		"code": "`%s`",
		"pre":  "```%s```",
		"a":    "[%s](%s)",
	}

	for _, i := range strings.Split(text, " ") {
		for htmlTag, keyValue := range Html2MdMap {
			k := ""
			// using this because <a> uses <href> tag
			if htmlTag == "a" {
				re := regexp.MustCompile(`<a href="(.*?)">(.*?)</a>`)
				if re.MatchString(i) {
					k = fmt.Sprintf(keyValue, re.FindStringSubmatch(i)[2], re.FindStringSubmatch(i)[1])
				} else {
					continue
				}
			} else {
				regexPattern := fmt.Sprintf(`<%s>(.*)<\/%s>`, htmlTag, htmlTag)
				pattern := regexp.MustCompile(regexPattern)
				if pattern.MatchString(i) {
					k = fmt.Sprintf(keyValue, pattern.ReplaceAllString(i, "$1"))
				} else {
					continue
				}
			}
			text = strings.ReplaceAll(text, i, k)
		}
	}

	return text
}
