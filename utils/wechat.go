package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/LIUHUANUCAS/auth/config"
)

const (
	// WeChatCode2SessionURL is the URL to exchange code for session info
	WeChatCode2SessionURL = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"
)

// WeChatManager handles WeChat API operations
type WeChatManager struct {
	config *config.WeChatConfig
	client *http.Client
}

// NewWeChatManager creates a new WeChatManager
func NewWeChatManager(config *config.WeChatConfig) *WeChatManager {
	return &WeChatManager{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Code2SessionResponse represents the response from the code2session API
type Code2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid,omitempty"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// Code2Session exchanges a code for session information
func (m *WeChatManager) Code2Session(code string) (*Code2SessionResponse, error) {
	// Build the URL
	url := fmt.Sprintf(WeChatCode2SessionURL, m.config.AppID, m.config.AppSecret, code)

	// Make the request
	resp, err := m.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to WeChat API: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the response
	var sessionResp Code2SessionResponse
	if err := json.Unmarshal(body, &sessionResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if sessionResp.ErrCode != 0 {
		return nil, fmt.Errorf("WeChat API error: %d - %s", sessionResp.ErrCode, sessionResp.ErrMsg)
	}

	return &sessionResp, nil
}
