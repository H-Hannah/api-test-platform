package config

import (
	"os"
	"strings"
)

type Config struct {
	Addr         string
	DBPath       string
	APIToken     string
	AIAPIKey     string
	AIBaseURL    string
	AIModel       string
	AIVisionModel string // 含设计稿截图时使用，如 qwen-vl-max
	DocsRepoRoot  string // 本地 Git 文档仓根目录，如 qa-doc-generator 路径
}

func Load() Config {
	loadDotEnv(".env")
	return Config{
		Addr:      getenv("ADDR", ":8080"),
		DBPath:    getenv("DB_PATH", "./data/platform.db"),
		APIToken:  getenv("API_TOKEN", "TEST123"),
		AIAPIKey:  getenv("AI_API_KEY", ""),
		AIBaseURL: getenv("AI_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"),
		AIModel:       getenv("AI_MODEL", "qwen-plus"),
		AIVisionModel: getenv("AI_VISION_MODEL", "qwen-vl-max"),
		DocsRepoRoot:  getenv("DOCS_REPO_ROOT", ""),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadDotEnv reads KEY=VALUE lines from path; existing env vars are not overwritten.
func loadDotEnv(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		os.Setenv(key, strings.TrimSpace(val))
	}
}
