package force

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	InstanceURL string `json:"instance_url"`
	Id          string `json:"id"`
	TokenType   string `json:"token_type"`
	IssuedAt    string `json:"issued_at"`
	Signature   string `json:"signature"`
}

type Client struct {
	httpClient http.Client
	auth       AuthResponse
	version    string
}

func (c *Client) AuthToken(instance, token string) (bool, error) {
	c.auth = AuthResponse{
		AccessToken: token,
		InstanceURL: instance,
	}

	return true, nil
}

func (c *Client) Auth(host, username, password, token, clientId, clientSecret string) (bool, error) {
	q := url.Values{}
	q.Add("grant_type", "password")
	q.Add("client_id", clientId)
	q.Add("client_secret", clientSecret)
	q.Add("username", username)
	q.Add("password", password+token)

	u := url.URL{}
	u.Scheme = "https"
	u.Host = host
	u.Path = "/services/oauth2/token"

	requestURL := u.String()
	fmt.Println("full url", requestURL)
	fmt.Println("params", q.Encode())

	req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(q.Encode()))
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		return false, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return false, nil
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return false, nil
	}

	a, err := parseAuth(body)
	if err != nil {
		return false, err
	}

	c.auth = a

	return true, nil
}

func parseAuth(b []byte) (AuthResponse, error) {
	a := AuthResponse{}
	err := json.Unmarshal(b, &a)
	return a, err
}

func NewClient() Client {
	return Client{
		httpClient: http.Client{
			Timeout: 5 * time.Second,
		},
		version: "v56.0",
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

func (c *Client) Post(path string, obj any) ([]byte, error) {
	requestURL := c.buildRequestURL(path)

	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	requestBody := bytes.NewReader(b)

	req, err := http.NewRequest(http.MethodPost, requestURL, requestBody)
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

func (c *Client) Get(path string) ([]byte, error) {
	b, _, e := c.GetWithHeaders(path)
	return b, e
}

func (c *Client) GetWithHeaders(path string) ([]byte, http.Header, error) {
	requestURL := c.buildRequestURL(path)

	fmt.Println("req:", requestURL)

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		return []byte{}, http.Header{}, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.auth.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return []byte{}, http.Header{}, nil
	}

	b, err := io.ReadAll(resp.Body)

	return b, resp.Header, err
}
