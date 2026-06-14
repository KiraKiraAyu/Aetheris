package delivery

import (
	"net/http"
	"time"
)

const (
	slackBodyTemplate    = `{"text":{{ quote .Text }}}`
	discordBodyTemplate  = `{"content":{{ quote .Text }}}`
	feishuBodyTemplate   = `{"msg_type":"text","content":{"text":{{ quote .Text }}}}`
	dingTalkBodyTemplate = `{"msgtype":"text","text":{"content":{{ quote .Text }}}}`
	weComBodyTemplate    = `{"msgtype":"text","text":{"content":{{ quote .Text }}}}`
)

func NewSlackProvider(config HTTPProviderConfig, client *http.Client) *HTTPProvider {
	return NewHTTPProvider(normalizeChatWebhookConfig("slack", config, slackBodyTemplate), client)
}

func NewDiscordProvider(config HTTPProviderConfig, client *http.Client) *HTTPProvider {
	return NewHTTPProvider(normalizeChatWebhookConfig("discord", config, discordBodyTemplate), client)
}

func NewFeishuProvider(config HTTPProviderConfig, client *http.Client) *HTTPProvider {
	return NewHTTPProvider(normalizeChatWebhookConfig("feishu", config, feishuBodyTemplate), client)
}

func NewDingTalkProvider(config HTTPProviderConfig, client *http.Client) *HTTPProvider {
	return NewHTTPProvider(normalizeChatWebhookConfig("dingtalk", config, dingTalkBodyTemplate), client)
}

func NewWeComProvider(config HTTPProviderConfig, client *http.Client) *HTTPProvider {
	return NewHTTPProvider(normalizeChatWebhookConfig("wecom", config, weComBodyTemplate), client)
}

func normalizeChatWebhookConfig(name string, config HTTPProviderConfig, defaultBodyTemplate string) HTTPProviderConfig {
	config.Name = name
	if config.Method == "" {
		config.Method = http.MethodPost
	}
	if config.BodyTemplate == "" {
		config.BodyTemplate = defaultBodyTemplate
	}
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}
	if config.SuccessStatusMin == 0 {
		config.SuccessStatusMin = 200
	}
	if config.SuccessStatusMax == 0 {
		config.SuccessStatusMax = 299
	}
	if config.Headers == nil {
		config.Headers = map[string]string{}
	}
	return config
}
