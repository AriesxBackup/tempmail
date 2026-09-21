package tempmail

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	APIAddress            = "https://api.smtp.dev"
	MercureAddress        = "https://mercure.smtp.dev/.well-known/mercure"
	DefaultTimeout        = 30 * time.Second
	DefaultUsernameLength = 10
	DefaultPasswordLength = 12
	PollInterval          = 2 * time.Second
	InboxPath             = "INBOX"

	contentTypeLD         = "application/ld+json"
	contentTypeMergePatch = "application/merge-patch+json"
)

type Client struct {
	apiAddress string
	apiKey     string
	httpClient *http.Client
	account    *Account
}

type ClientOption func(*Client)

func WithAPIAddress(url string) ClientOption {
	return func(c *Client) {
		c.apiAddress = strings.TrimRight(url, "/")
	}
}

func WithAPIKey(key string) ClientOption {
	return func(c *Client) {
		c.apiKey = key
	}
}

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		apiAddress: APIAddress,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) send(method, path string, query url.Values, body any, accept string) (*http.Response, error) {
	if c.apiKey == "" {
		return nil, ErrAPIKeyRequired
	}

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}

	target := c.apiAddress + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}

	req, err := http.NewRequest(method, target, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("Accept", accept)
	if body != nil {
		if method == http.MethodPatch {
			req.Header.Set("Content-Type", contentTypeMergePatch)
		} else {
			req.Header.Set("Content-Type", contentTypeLD)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		defer resp.Body.Close()
		return nil, newAPIError(resp)
	}

	return resp, nil
}

func (c *Client) do(method, path string, query url.Values, body any, out any) error {
	resp, err := c.send(method, path, query, body, contentTypeLD)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if out == nil || resp.StatusCode == http.StatusNoContent {
		io.Copy(io.Discard, resp.Body)
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) fetchBytes(path string) ([]byte, error) {
	resp, err := c.send(http.MethodGet, path, nil, nil, "*/*")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func list[T any](c *Client, path string, query url.Values) ([]T, error) {
	var col collection[T]
	if err := c.do(http.MethodGet, path, query, nil, &col); err != nil {
		return nil, err
	}
	return col.Member, nil
}

func newAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	var problem struct {
		Message string `json:"message"`
		Detail  string `json:"detail"`
		Title   string `json:"title"`
	}
	msg := strings.TrimSpace(string(body))
	if json.Unmarshal(body, &problem) == nil {
		switch {
		case problem.Detail != "":
			msg = problem.Detail
		case problem.Message != "":
			msg = problem.Message
		case problem.Title != "":
			msg = problem.Title
		}
	}

	return &APIError{StatusCode: resp.StatusCode, Message: msg}
}

func pageQuery(page int) url.Values {
	q := url.Values{}
	if page > 1 {
		q.Set("page", itoa(page))
	}
	return q
}

func (c *Client) GetAccount() *Account {
	return c.account
}

func (c *Client) SetAccount(account *Account) {
	c.account = account
}

func (c *Client) HasAPIKey() bool {
	return c.apiKey != ""
}
