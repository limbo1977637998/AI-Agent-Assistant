package httpclient

import (
	"net/http"
	"net/url"
	"os"
	"time"

	aiagentconfig "ai-agent-assistant/internal/config"
)

var globalConfig *aiagentconfig.ProxyConfig

// SetConfig 设置全局代理配置
func SetConfig(cfg *aiagentconfig.ProxyConfig) {
	globalConfig = cfg

	// 如果启用了代理，同时设置环境变量
	if cfg.Enabled {
		if cfg.HTTPProxy != "" {
			os.Setenv("HTTP_PROXY", cfg.HTTPProxy)
			os.Setenv("http_proxy", cfg.HTTPProxy)
		}
		if cfg.HTTPSProxy != "" {
			os.Setenv("HTTPS_PROXY", cfg.HTTPSProxy)
			os.Setenv("https_proxy", cfg.HTTPSProxy)
		}
		if cfg.NoProxy != "" {
			os.Setenv("NO_PROXY", cfg.NoProxy)
			os.Setenv("no_proxy", cfg.NoProxy)
		}
	}
}

// NewClient 创建HTTP客户端，支持代理
func NewClient() *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	// 如果配置了代理，强制使用
	if globalConfig != nil && globalConfig.Enabled {
		proxyURL := globalConfig.HTTPSProxy
		if proxyURL == "" {
			proxyURL = globalConfig.HTTPProxy
		}

		if proxyURL != "" {
			proxy, _ := url.Parse(proxyURL)
			transport.Proxy = http.ProxyURL(proxy)
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

// GetProxyURL 获取代理URL（用于显示）
func GetProxyURL() string {
	if globalConfig != nil && globalConfig.Enabled {
		if globalConfig.HTTPSProxy != "" {
			return globalConfig.HTTPSProxy
		}
		if globalConfig.HTTPProxy != "" {
			return globalConfig.HTTPProxy
		}
	}
	return ""
}

// GetProxyInfo 获取代理信息（用于日志）
func GetProxyInfo() string {
	if globalConfig != nil && globalConfig.Enabled {
		return globalConfig.HTTPSProxy
	}
	return "未配置代理"
}
