package tempmail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	APIAddress            = "https://api.mail.tm"
	DefaultTimeout        = 30 * time.Second
	DefaultUsernameLength = 10
	DefaultPasswordLength = 6
	PollInterval          = 2 * time.Second
)

type Client struct {
	apiAddress  string
	httpClient  *http.Client
	account     *Account
	token       string
	authHeaders map[string]string
}

type ClientOption func(*Client)

func WithAPIAddress(url string) ClientOption {
	return func(c *Client) {
		c.apiAddress = url
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
		authHeaders: make(map[string]string),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) GetDomains() ([]string, error) {
	for i := 0; i < 3; i++ {
		resp, err := c.httpClient.Get(c.apiAddress + "/domains")
		if err != nil {
			time.Sleep(PollInterval)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var domainResp DomainResponse
			if err := json.NewDecoder(resp.Body).Decode(&domainResp); err != nil {
				return nil, fmt.Errorf("decode failed: %w", err)
			}

			domains := make([]string, len(domainResp.Member))
			for i, d := range domainResp.Member {
				domains[i] = d.Domain
			}
			return domains, nil
		}
		time.Sleep(PollInterval)
	}
	return nil, ErrNoDomains
}

func (c *Client) CreateAccount(password string) (*Account, error) {
	domains, err := c.GetDomains()
	if err != nil {
		return nil, err
	}
	if len(domains) == 0 {
		return nil, ErrNoDomains
	}

	username := GenerateUsername(DefaultUsernameLength)
	address := fmt.Sprintf("%s@%s", username, domains[0])

	if password == "" {
		password = GeneratePassword(DefaultPasswordLength)
	}

	account, err := c.makeAccountRequest("accounts", address, password)
	if err != nil {
		return nil, err
	}

	account.Password = password
	c.account = account

	if err := c.Login(address, password); err != nil {
		return nil, err
	}

	return account, nil
}

func (c *Client) makeAccountRequest(endpoint, address, password string) (*Account, error) {
	payload := map[string]string{
		"address":  address,
		"password": password,
	}

	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, c.apiAddress+"/"+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/ld+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: HTTP %d - %s", ErrCouldNotGetAccount, resp.StatusCode, string(body))
	}

	var account Account
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, err
	}

	return &account, nil
}

func (c *Client) Login(address, password string) error {
	if address == "" && c.account != nil {
		address = c.account.Address
		password = c.account.Password
	}

	if address == "" || password == "" {
		return ErrAddressRequired
	}

	payload := map[string]string{
		"address":  address,
		"password": password,
	}

	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, c.apiAddress+"/token", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/ld+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: HTTP %d - %s", ErrCouldNotGetAccount, resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	c.token = tokenResp.Token
	c.authHeaders = map[string]string{
		"Accept":        "application/ld+json",
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + c.token,
	}

	return nil
}

func (c *Client) GetMessages(page int) ([]Message, error) {
	if c.token == "" {
		return nil, ErrNotAuthenticated
	}

	if page < 1 {
		page = 1
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/messages?page=%d", c.apiAddress, page), nil)
	if err != nil {
		return nil, err
	}
	for k, v := range c.authHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrCouldNotGetMessages, resp.StatusCode)
	}

	var hydraResp HydraResponse
	if err := json.NewDecoder(resp.Body).Decode(&hydraResp); err != nil {
		return nil, err
	}

	messages := make([]Message, 0, len(hydraResp.Member))
	for _, m := range hydraResp.Member {
		time.Sleep(PollInterval)

		fullMsg, err := c.GetMessage(m.ID)
		if err != nil {
			continue
		}
		messages = append(messages, *fullMsg)
	}

	return messages, nil
}

func (c *Client) GetMessage(messageID string) (*Message, error) {
	if c.token == "" {
		return nil, ErrNotAuthenticated
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/messages/%s", c.apiAddress, messageID), nil)
	if err != nil {
		return nil, err
	}
	for k, v := range c.authHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrCouldNotGetMessages, resp.StatusCode)
	}

	var message Message
	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		return nil, err
	}

	return &message, nil
}

func (c *Client) MarkMessageSeen(messageID string) error {
	if c.token == "" {
		return ErrNotAuthenticated
	}

	payload := map[string]bool{"seen": true}
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/messages/%s", c.apiAddress, messageID), bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/ld+json")
	req.Header.Set("Content-Type", "application/merge-patch+json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("%w: HTTP %d", ErrCouldNotUpdateMessage, resp.StatusCode)
	}

	return nil
}

func (c *Client) DeleteMessage(messageID string) error {
	if c.token == "" {
		return ErrNotAuthenticated
	}

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/messages/%s", c.apiAddress, messageID), nil)
	if err != nil {
		return err
	}
	for k, v := range c.authHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete failed: HTTP %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) DeleteAccount() error {
	if c.token == "" || c.account == nil {
		return ErrNoActiveAccount
	}

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/accounts/%s", c.apiAddress, c.account.ID), nil)
	if err != nil {
		return err
	}
	for k, v := range c.authHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete failed: HTTP %d", resp.StatusCode)
	}

	c.account = nil
	c.token = ""
	c.authHeaders = make(map[string]string)
	return nil
}

func (c *Client) WaitForMessage() (*Message, error) {
	if c.token == "" {
		return nil, ErrNotAuthenticated
	}

	oldIDs := make(map[string]bool)
	for {
		messages, err := c.GetMessages(1)
		if err == nil {
			for _, m := range messages {
				oldIDs[m.ID] = true
			}
			break
		}
		time.Sleep(3 * time.Second)
	}

	for {
		time.Sleep(PollInterval)
		messages, err := c.GetMessages(1)
		if err != nil {
			continue
		}

		for _, m := range messages {
			if !oldIDs[m.ID] {
				return &m, nil
			}
		}
	}
}

func (c *Client) GetAccount() *Account {
	return c.account
}

func (c *Client) IsAuthenticated() bool {
	return c.token != ""
}
