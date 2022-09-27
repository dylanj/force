package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/davecgh/go-spew/spew"
)

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	InstanceURL string `json:"instance_url"`
	Id          string `json:"id"`
	TokenType   string `json:"token_type"`
	IssuedAt    string `json:"issued_at"`
	Signature   string `json:"signature"`
}

type QueryJobBody struct {
	// Operation can be either query (non deleted) or queryAll (includes deleted
	// records)
	Operation       string `json:"operation"`
	Query           string `json:"query"`
	ContentType     string `json:"contentType,omitempty"`
	ColumnDelimeter string `json:"columnDelimeter,omitempty"`
	LineEnding      string `json:"lineEnding,omitempty"`
}

type QueryJobResponse struct {
	Id              string  `json:"id"`
	Operation       string  `json:"operation"`
	Object          string  `json:"object"`
	CreatedById     string  `json:"createdById"`
	CreatedDate     string  `json:"createdDate"`
	SystemModstamp  string  `json:"systemModstamp"`
	State           string  `json:"state"`
	ConcurrencyMode string  `json:"concurrencyMode"`
	ContentType     string  `json:"content_type"`
	APIVersion      float64 `json:"apiVersion"`
	LineEnding      string  `json:"lineEnding"`
	ColumnDelimiter string  `json:"columnDelimiter"`
}

type Client struct {
	httpClient http.Client
	auth       AuthResponse
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
	}
}

func (c Client) buildRequestURL(path string) string {
	return c.auth.InstanceURL + path
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

func main() {
	c := NewClient()
	/*
		salesforce_host := os.Getenv("SALESFORCE_HOST")
		salesforce_username := os.Getenv("SALESFORCE_USERNAME")
		salesforce_password := os.Getenv("SALESFORCE_PASSWORD")
		salesforce_token := os.Getenv("SALESFORCE_TOKEN")
		salesforce_client_id := os.Getenv("SALESFORCE_CLIENT_ID")
		salesforce_client_secret := os.Getenv("SALESFORCE_CLIENT_SECRET")

		ok, err := c.Auth(
			salesforce_host,
			salesforce_username,
			salesforce_password,
			salesforce_token,
			salesforce_client_id,
			salesforce_client_secret,
		)
	*/

	token := os.Getenv("TOKEN")
	instance := os.Getenv("INSTANCE")

	ok, err := c.AuthToken(instance, token)
	if err != nil {
		os.Exit(0)
	}

	if ok {
		fmt.Println("auth success")
	} else {
		fmt.Println("auth no beuno")
	}

	q, err := c.QueryJob("SELECT Id, Name FROM Account LIMIT 10")

	for {
		q, err := c.QueryJobStatus(q.Id)
		if err != nil {
			fmt.Println("got an error")
			spew.Dump(err)
			os.Exit(1)
		}

		if q.State == "JobComplete" {
			fmt.Println("finished")
			break
		}

		fmt.Println("not finished")
	}

	c.QueryJobResults(q.Id)
}

func (c *Client) QueryJobResults(jobId string) string {
	q := url.Values{}
	q.Add("maxRecords", "2")

	for {
		path := "/services/data/v55.0/jobs/query/" + jobId + "/results?" + q.Encode()
		b, h, err := c.GetWithHeaders(path)
		if err != nil {
			fmt.Println("err")
			spew.Dump(err)
		}
		fmt.Println(string(b))

		locator := h.Get("Sforce-Locator")
		q.Set("locator", locator)

		if locator == "null" {
			fmt.Println("breaking")
			break
		}
	}

	return ""
}

func (c *Client) QueryJobStatus(jobId string) (*QueryJobResponse, error) {
	b, err := c.Get("/services/data/v55.0/jobs/query/" + jobId)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(b))

	resp := QueryJobResponse{}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) QueryJob(query string) (*QueryJobResponse, error) {
	q := QueryJobBody{
		Operation: "query",
		Query:     query,
	}

	fmt.Println("creating query")
	b, err := c.Post("/services/data/v55.0/jobs/query", q)
	if err != nil {
		return nil, err
	}

	resp := QueryJobResponse{}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
