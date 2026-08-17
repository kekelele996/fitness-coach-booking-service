package config

import (
	"os"
	"strconv"
	"strings"
)

// Config 运行期配置。
type Config struct {
	Workers        int
	RetryLimit     int
	ExpireMinutes  int
	PollIntervalMs int
	SkillRoutes    map[string]string
}

func defaultSkillRoutes() map[string]string {
	return map[string]string{
		"strength": "strength-cert",
		"yoga":     "yoga-cert",
		"hiit":     "hiit-cert",
		"default":  "general",
	}
}

func loadSkillRoutes() map[string]string {
	v := os.Getenv("FITNESS_SKILL_ROUTES")
	if v == "" {
		return defaultSkillRoutes()
	}
	routes := map[string]string{}
	for _, pair := range strings.Split(v, ",") {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) == 2 {
			routes[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	if len(routes) == 0 {
		return nil
	}
	return routes
}

// Load 从环境变量加载配置。
func Load() *Config {
	return &Config{
		Workers:        getInt("FITNESS_WORKERS", 4),
		RetryLimit:     getInt("FITNESS_RETRY_LIMIT", 3),
		ExpireMinutes:  getInt("FITNESS_EXPIRE_MINUTES", 30),
		PollIntervalMs: getInt("FITNESS_POLL_INTERVAL_MS", 500),
		SkillRoutes:    loadSkillRoutes(),
	}
}

func getInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}
