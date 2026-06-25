package impact

import (
	"strings"
	"unicode"
)

// keywordsFromDescription 从口述/文本变更提取可匹配用例的关键词。
func keywordsFromDescription(text string) []string {
	text = strings.ToLower(text)
	seen := map[string]bool{}
	var keys []string
	add := func(k string) {
		k = strings.Trim(k, "：:、；;\"'（）()[]<>")
		if len(k) < 2 || seen[k] {
			return
		}
		seen[k] = true
		keys = append(keys, k)
	}
	for _, m := range apiPathRe.FindAllStringSubmatch(text, -1) {
		if len(m) >= 2 {
			add(m[1])
		}
	}
	for _, m := range routeRe.FindAllStringSubmatch(text, -1) {
		if len(m) >= 3 {
			add(m[2])
		}
	}
	var tok strings.Builder
	flush := func() {
		if tok.Len() > 0 {
			add(tok.String())
			tok.Reset()
		}
	}
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			tok.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return keys
}

func isDescriptionSource(files []prFile) bool {
	return len(files) == 1 && files[0].Status == "description"
}
