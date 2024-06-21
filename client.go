package force

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/net/publicsuffix"
)

type Client struct {
	httpClient *http.Client
	auth       *AuthResponse
	version    string
}

func (c *Client) Auth(auth authenticator) error {
	ar, err := auth.Authenticate()
	if err != nil {
		return err
	}

	c.auth = ar

	return nil
}

func (c *Client) DebugAuth() {
	spew.Dump(c.auth)
}

func NewClient() Client {
	cookiejarOptions := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, _ := cookiejar.New(&cookiejarOptions)

	return Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
		version: "56.0",
	}
}

func (c Client) buildRequestURL(path string) string {
	return c.auth.InstanceURL + path
}

func (c *Client) Patch(path string, obj any) ([]byte, error) {
	requestURL := c.buildRequestURL(path)

	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	requestBody := bytes.NewReader(b)

	req, err := http.NewRequest(http.MethodPatch, requestURL, requestBody)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		return []byte{}, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.auth.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return []byte{}, nil
	}

	return io.ReadAll(resp.Body)
}

// todo: extract auth into own struct, pass that around
func (c *Client) postWithClient(httpClient *http.Client, path string, obj any) ([]byte, error) {
	requestURL := c.buildRequestURL(path)

	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	requestBody := bytes.NewReader(b)

	req, err := http.NewRequest(http.MethodPost, requestURL, requestBody)
	if err != nil {
		return []byte{}, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.auth.AccessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return []byte{}, err
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) Post(path string, obj any) ([]byte, error) {
	return c.postWithClient(c.httpClient, path, obj)
}

func (c *Client) Get(path string) ([]byte, error) {
	b, _, e := c.GetWithHeaders(path)
	return b, e
}

func (c *Client) GetWithHeaders(path string) ([]byte, http.Header, error) {
	requestURL := c.buildRequestURL(path)

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return []byte{}, http.Header{}, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.auth.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return []byte{}, http.Header{}, err
	}

	b, err := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return b, resp.Header, err
	}

	return b, resp.Header, errors.New(string(b))
}
