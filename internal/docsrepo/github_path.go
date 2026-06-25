package docsrepo

import (
	"net/url"
	"regexp"
	"strings"
)

// githubBlobRe 匹配 github.com/<owner>/<repo>/blob/<ref>/<path>
var githubBlobRe = regexp.MustCompile(
	`^https?://github\.com/([^/]+/[^/]+)/blob/([^/]+)/(.+?)(?:\?.*)?$`,
)

// ResolvePackagePath 将 GitHub blob 链接或相对路径统一为文档仓内相对路径。
func ResolvePackagePath(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}
	if m := githubBlobRe.FindStringSubmatch(input); m != nil {
		p, err := url.PathUnescape(m[3])
		if err == nil {
			return strings.TrimPrefix(p, "/")
		}
		return strings.TrimPrefix(m[3], "/")
	}
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		return ""
	}
	return strings.Trim(strings.TrimPrefix(input, "/"), " ")
}
