package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Options 请求选项
type Options struct {
	Params         map[string]string // URL查询参数
	Headers        map[string]string // 请求头
	Cookies        map[string]string // Cookie
	Timeout        time.Duration     // 超时时间
	AllowRedirects *bool             // 是否允许重定向
}

// Get 发送GET请求
func (c *Client) Get(urlStr string, opts *Options) (*Response, error) {
	return c.doRequest("GET", urlStr, nil, opts)
}

// Post 发送POST请求
func (c *Client) Post(urlStr string, body interface{}, opts *Options) (*Response, error) {
	return c.doRequest("POST", urlStr, body, opts)
}

// Put 发送PUT请求
func (c *Client) Put(urlStr string, body interface{}, opts *Options) (*Response, error) {
	return c.doRequest("PUT", urlStr, body, opts)
}

// Delete 发送DELETE请求
func (c *Client) Delete(urlStr string, opts *Options) (*Response, error) {
	return c.doRequest("DELETE", urlStr, nil, opts)
}

// Patch 发送PATCH请求
func (c *Client) Patch(urlStr string, body interface{}, opts *Options) (*Response, error) {
	return c.doRequest("PATCH", urlStr, body, opts)
}

// Head 发送HEAD请求
func (c *Client) Head(urlStr string, opts *Options) (*Response, error) {
	return c.doRequest("HEAD", urlStr, nil, opts)
}

// Options 发送OPTIONS请求
func (c *Client) Options(urlStr string, opts *Options) (*Response, error) {
	return c.doRequest("OPTIONS", urlStr, nil, opts)
}

// PostJSON 发送JSON数据
func (c *Client) PostJSON(urlStr string, data interface{}, opts *Options) (*Response, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %w", err)
	}

	if opts == nil {
		opts = &Options{}
	}
	if opts.Headers == nil {
		opts.Headers = make(map[string]string)
	}
	opts.Headers["Content-Type"] = "application/json"

	return c.doRequest("POST", urlStr, jsonBytes, opts)
}

// PostForm 发送表单数据
func (c *Client) PostForm(urlStr string, data map[string]string, opts *Options) (*Response, error) {
	formData := make(url.Values)
	for k, v := range data {
		formData.Set(k, v)
	}

	if opts == nil {
		opts = &Options{}
	}
	if opts.Headers == nil {
		opts.Headers = make(map[string]string)
	}
	opts.Headers["Content-Type"] = "application/x-www-form-urlencoded"

	return c.doRequest("POST", urlStr, []byte(formData.Encode()), opts)
}

// PostBytes 发送字节数据
func (c *Client) PostBytes(urlStr string, data []byte, opts *Options) (*Response, error) {
	if opts == nil {
		opts = &Options{}
	}
	if opts.Headers == nil {
		opts.Headers = make(map[string]string)
	}
	if _, ok := opts.Headers["Content-Type"]; !ok {
		opts.Headers["Content-Type"] = "application/octet-stream"
	}

	return c.doRequest("POST", urlStr, data, opts)
}

// FileField 文件字段定义
type FileField struct {
	FieldName   string // 表单字段名
	FileName    string // 文件名
	ContentType string // MIME类型（可选）
	FilePath    string // 本地文件路径（与Data二选一）
	Data        []byte // 文件内容（与FilePath二选一）
}

// PostMultipart 发送multipart表单数据
func (c *Client) PostMultipart(urlStr string, fields map[string]string, files []FileField, opts *Options) (*Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加普通字段
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("写入字段失败: %w", err)
		}
	}

	// 添加文件
	for _, file := range files {
		var fileContent []byte
		var err error

		if file.FilePath != "" {
			fileContent, err = os.ReadFile(file.FilePath)
			if err != nil {
				return nil, fmt.Errorf("读取文件失败: %w", err)
			}
			if file.FileName == "" {
				file.FileName = filepath.Base(file.FilePath)
			}
		} else {
			fileContent = file.Data
		}

		if file.FieldName == "" {
			file.FieldName = "file"
		}
		if file.FileName == "" {
			file.FileName = "file"
		}

		part, err := writer.CreateFormFile(file.FieldName, file.FileName)
		if err != nil {
			return nil, fmt.Errorf("创建文件字段失败: %w", err)
		}

		if _, err := part.Write(fileContent); err != nil {
			return nil, fmt.Errorf("写入文件内容失败: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("关闭multipart失败: %w", err)
	}

	if opts == nil {
		opts = &Options{}
	}
	if opts.Headers == nil {
		opts.Headers = make(map[string]string)
	}
	opts.Headers["Content-Type"] = writer.FormDataContentType()

	return c.doRequest("POST", urlStr, body.Bytes(), opts)
}

// PostFile 上传单个文件
func (c *Client) PostFile(urlStr string, fieldName string, filePath string, opts *Options) (*Response, error) {
	return c.PostMultipart(urlStr, nil, []FileField{
		{FieldName: fieldName, FilePath: filePath},
	}, opts)
}

// doRequest 执行HTTP请求
func (c *Client) doRequest(method, urlStr string, body interface{}, opts *Options) (*Response, error) {
	if opts == nil {
		opts = &Options{}
	}

	// 构建URL参数
	if opts.Params != nil {
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			return nil, fmt.Errorf("解析URL失败: %w", err)
		}
		query := parsedURL.Query()
		for k, v := range opts.Params {
			query.Set(k, v)
		}
		parsedURL.RawQuery = query.Encode()
		urlStr = parsedURL.String()
	}

	// 构建请求体
	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case []byte:
			bodyReader = bytes.NewReader(v)
		case string:
			bodyReader = strings.NewReader(v)
		case io.Reader:
			bodyReader = v
		default:
			// 尝试JSON序列化
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("序列化请求体失败: %w", err)
			}
			bodyReader = bytes.NewReader(jsonBytes)
		}
	}

	// 创建请求
	req, err := http.NewRequest(method, urlStr, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置默认headers
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// 设置请求headers
	if opts.Headers != nil {
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}
	}

	// 设置cookies
	for k, v := range c.cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
	}
	if opts.Cookies != nil {
		for k, v := range opts.Cookies {
			req.AddCookie(&http.Cookie{Name: k, Value: v})
		}
	}

	// 保存原始配置
	originalTimeout := c.httpClient.Timeout
	originalRedirect := c.httpClient.CheckRedirect

	// 临时修改配置
	if opts.Timeout > 0 {
		c.httpClient.Timeout = opts.Timeout
	}

	if opts.AllowRedirects != nil && !*opts.AllowRedirects {
		c.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// 发送请求
	resp, err := c.httpClient.Do(req)

	// 恢复配置
	c.httpClient.Timeout = originalTimeout
	c.httpClient.CheckRedirect = originalRedirect

	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 更新cookies
	for _, cookie := range resp.Cookies() {
		c.cookies[cookie.Name] = cookie.Value
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header,
		Cookies:    resp.Cookies(),
		Body:       respBody,
		Request:    req,
	}, nil
}
