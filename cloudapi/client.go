package cloudapi

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/Drunkard-baifeng/golibs/httpclient"
	"github.com/Drunkard-baifeng/golibs/logger"
)

const (
	DefaultBaseURL = "http://127.0.0.1:8081"
)

// Client 云端API客户端
type Client struct {
	client  *httpclient.Client
	baseURL string
	key     string // uid，登录后获取
}

var (
	instance *Client
	once     sync.Once
)

// Default 获取单例实例（懒加载）
func Default() *Client {
	once.Do(func() {
		instance = &Client{
			client:  httpclient.New(),
			baseURL: DefaultBaseURL,
		}
		instance.client.UpdateHeaders(map[string]string{
			"Content-Type": "application/json",
		})
	})
	return instance
}

// New 创建新实例（非单例）
func New(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	c := &Client{
		client:  httpclient.New(),
		baseURL: baseURL,
	}
	c.client.UpdateHeaders(map[string]string{
		"Content-Type": "application/json",
	})
	return c
}

// SetBaseURL 设置服务器地址
func (c *Client) SetBaseURL(baseURL string) *Client {
	c.baseURL = baseURL
	return c
}

// SetKey 设置用户标识
func (c *Client) SetKey(key string) *Client {
	c.key = key
	return c
}

// GetKey 获取用户标识
func (c *Client) GetKey() string {
	return c.key
}

// SetProxy 设置代理
// proxyType: "http" 或 "socks5"
func (c *Client) SetProxy(proxy, proxyType string) *Client {
	c.client.SetProxy(proxy, proxyType)
	return c
}

// buildURL 构建带 key 参数的 URL
func (c *Client) buildURL(path string) string {
	if c.key != "" {
		return fmt.Sprintf("%s%s?key=%s", c.baseURL, path, c.key)
	}
	return fmt.Sprintf("%s%s", c.baseURL, path)
}

// getCallerName 获取调用者函数名
func getCallerName() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	name := fn.Name()
	// 取最后一个.后面的部分
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		name = name[idx+1:]
	}
	return name
}

// request 带重试的请求方法
func (c *Client) request(method, path string, body interface{}, maxRetries int) (*Response, error) {
	if maxRetries <= 0 {
		maxRetries = 2
	}

	funcName := getCallerName()
	var lastErr error

	for retry := 0; retry < maxRetries; retry++ {
		var resp *httpclient.Response
		var err error

		url := c.buildURL(path)
		// logger.Infof("请求URL: %s", url)
		// logger.Infof("请求方法: %s", method)
		// logger.Infof("请求体: %v", body)

		if method == "GET" {
			resp, err = c.client.Get(url, nil)
		} else {
			resp, err = c.client.Post(url, body, nil)
		}

		if err != nil {
			lastErr = err
			logger.Errorf("%s 第%d次重试异常: %v", funcName, retry+1, err)
			continue
		}

		var result Response
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			lastErr = fmt.Errorf("解析响应失败: %w, body: %s", err, resp.Text())
			logger.Errorf("%s 第%d次重试解析失败: %v", funcName, retry+1, lastErr)
			continue
		}

		if result.Code == 200 {
			logger.Successf("%s 成功, 重试次数:%d, 结果:%s", funcName, retry+1, resp.Text())
			return &result, nil
		}

		logger.Errorf("%s 失败, 重试次数:%d, 结果:%s", funcName, retry+1, resp.Text())
		return &result, fmt.Errorf(result.Msg)
	}

	return nil, lastErr
}

// doRequest 执行请求（兼容旧接口）
func (c *Client) doRequest(path string, body interface{}) (*Response, error) {
	return c.request("POST", path, body, 2)
}

// ==================== 登录 ====================

// Login 云端登录
func (c *Client) Login(username, password string) error {
	resp, err := c.request("POST", "/api/user/login", map[string]string{
		"username": username,
		"password": password,
	}, 2)
	if err != nil {
		return err
	}

	logger.Infof("登录响应: %s", resp.Msg)

	// 解析 token 获取 uid
	if data, ok := resp.Data.(map[string]interface{}); ok {
		if token, ok := data["token"].(string); ok {
			uid, err := parseJWTUID(token)
			if err != nil {
				return fmt.Errorf("解析token失败: %w", err)
			}
			c.key = uid
			return nil
		}
	}

	return fmt.Errorf("登录响应格式错误")
}

// parseJWTUID 从 JWT token 中解析 uid（不验证签名）
func parseJWTUID(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}

	// 解码 payload（第二部分）
	payload := parts[1]
	// 补齐 base64 padding
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		// 尝试标准 base64
		decoded, err = base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return "", fmt.Errorf("decode payload failed: %w", err)
		}
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return "", fmt.Errorf("parse claims failed: %w", err)
	}

	if uid, ok := claims["uid"].(string); ok {
		return uid, nil
	}

	return "", fmt.Errorf("uid not found in token")
}

// ==================== 类别操作 ====================

// ConfigPost 添加养号类别
func (c *Client) ConfigPost(mode string) (uint, error) {
	resp, err := c.doRequest("/api/number_maintenance/task/mode/post", ConfigPostReq{Mode: mode})
	if err != nil {
		return 0, err
	}
	if resp.Code != 200 {
		return 0, fmt.Errorf(resp.Msg)
	}

	if data, ok := resp.Data.(map[string]interface{}); ok {
		if id, ok := data["id"].(float64); ok {
			return uint(id), nil
		}
	}
	return 0, nil
}

// ==================== 数据操作 ====================

// DataPost 添加养号数据
func (c *Client) DataPost(req *DataPostReq) (uint, error) {
	resp, err := c.doRequest("/api/number_maintenance/task/data/post", req)
	if err != nil {
		return 0, err
	}
	if resp.Code != 200 {
		return 0, fmt.Errorf(resp.Msg)
	}

	if data, ok := resp.Data.(map[string]interface{}); ok {
		if id, ok := data["id"].(float64); ok {
			return uint(id), nil
		}
	}
	return 0, nil
}

// DataGet 获取养号数据
func (c *Client) DataGet(configID uint, nextTimeMode string) (*DataGetResp, error) {
	resp, err := c.doRequest("/api/number_maintenance/task/data/get", DataGetReq{
		ConfigID:     configID,
		NextTimeMode: nextTimeMode,
	})
	if err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf(resp.Msg)
	}

	// 解析 data 字段
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var result DataGetResp
	if err := json.Unmarshal(dataBytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DataSet 修改养号数据
func (c *Client) DataSet(req *DataSetReq) error {
	resp, err := c.doRequest("/api/number_maintenance/task/data/set", req)
	if err != nil {
		return err
	}
	if resp.Code != 200 {
		return fmt.Errorf(resp.Msg)
	}
	return nil
}

// DataDel 删除养号数据
func (c *Client) DataDel(id uint) error {
	resp, err := c.doRequest("/api/number_maintenance/task/data/del", DataDelReq{ID: id})
	if err != nil {
		return err
	}
	if resp.Code != 200 {
		return fmt.Errorf(resp.Msg)
	}
	return nil
}

// ==================== 重置操作 ====================

// ResettingAll 重置全部养号任务
func (c *Client) ResettingAll(configID uint) (int64, error) {
	resp, err := c.doRequest("/api/number_maintenance/task/resetting/all", ResettingAllReq{ConfigID: configID})
	if err != nil {
		return 0, err
	}
	if resp.Code != 200 {
		return 0, fmt.Errorf(resp.Msg)
	}

	if data, ok := resp.Data.(map[string]interface{}); ok {
		if count, ok := data["count"].(float64); ok {
			return int64(count), nil
		}
	}
	return 0, nil
}

// ResettingOne 重置单条养号任务
func (c *Client) ResettingOne(id uint) error {
	resp, err := c.doRequest("/api/number_maintenance/task/resetting/one", ResettingOneReq{ID: id})
	if err != nil {
		return err
	}
	if resp.Code != 200 {
		return fmt.Errorf(resp.Msg)
	}
	return nil
}

// ==================== 日志操作 ====================

// TimeLogGet 查询养号日志（单条）
func (c *Client) TimeLogGet(id uint, date string) (string, error) {
	datePtr := &date
	if date == "" {
		datePtr = nil
	}

	resp, err := c.doRequest("/api/number_maintenance/task/time_log/get", TimeLogGetReq{
		ID:   id,
		Date: datePtr,
	})
	if err != nil {
		return "", err
	}
	if resp.Code != 200 {
		return "", fmt.Errorf(resp.Msg)
	}

	if data, ok := resp.Data.(map[string]interface{}); ok {
		if content, ok := data["content"].(string); ok {
			return content, nil
		}
	}
	return "", nil
}

// TimeLogGetAll 查询全部养号日志
func (c *Client) TimeLogGetAll(id uint) (map[string]string, error) {
	resp, err := c.doRequest("/api/number_maintenance/task/time_log/get", TimeLogGetReq{ID: id})
	if err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf(resp.Msg)
	}

	result := make(map[string]string)
	if data, ok := resp.Data.(map[string]interface{}); ok {
		if logs, ok := data["logs"].(map[string]interface{}); ok {
			for k, v := range logs {
				if str, ok := v.(string); ok {
					result[k] = str
				}
			}
		}
	}
	return result, nil
}

// TimeLogPost 添加/修改养号日志
func (c *Client) TimeLogPost(id uint, date, content string) error {
	var datePtr, contentPtr *string
	if date != "" {
		datePtr = &date
	}
	if content != "" {
		contentPtr = &content
	}

	resp, err := c.doRequest("/api/number_maintenance/task/time_log/post", TimeLogPostReq{
		ID:      id,
		Date:    datePtr,
		Content: contentPtr,
	})
	if err != nil {
		return err
	}
	if resp.Code != 200 {
		return fmt.Errorf(resp.Msg)
	}
	return nil
}

// TimeLogDel 删除养号日志
func (c *Client) TimeLogDel(id uint, date string) error {
	resp, err := c.doRequest("/api/number_maintenance/task/time_log/del", TimeLogDelReq{
		ID:   id,
		Date: date,
	})
	if err != nil {
		return err
	}
	if resp.Code != 200 {
		return fmt.Errorf(resp.Msg)
	}
	return nil
}
