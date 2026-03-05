package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/viper"
)

const baseURL = "https://api.gumlet.com/v1"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient() (*Client, error) {
	apiKey := viper.GetString("api-key")
	if apiKey == "" {
		return nil, fmt.Errorf("API key is not set. Please run 'gumlet login' to configure your API key")
	}
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}, nil
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	url := baseURL + path
	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	return req, nil
}

func (c *Client) Do(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	return body, nil
}

func (c *Client) Get(path string, queryParams map[string]string) ([]byte, error) {
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for k, v := range queryParams {
		if v != "" {
			q.Add(k, v)
		}
	}
	req.URL.RawQuery = q.Encode()

	return c.Do(req)
}

func (c *Client) Post(path string, body interface{}) ([]byte, error) {
	req, err := c.newRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Delete(path string) ([]byte, error) {
	req, err := c.newRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) DeleteWithBody(path string, body interface{}) ([]byte, error) {
	req, err := c.newRequest("DELETE", path, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Put(path string, body interface{}) ([]byte, error) {
	req, err := c.newRequest("PUT", path, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
