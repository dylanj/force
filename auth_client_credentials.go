package force

import (
	b64 "encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ClientCredentialAuth struct {
	host         string
	clientId     string
	clientSecret string
}

func AuthClientCredentials(host, clientId, clientSecret string) (*ClientCredentialAuth, error) {
	a := ClientCredentialAuth{
		host:         host,
		clientId:     clientId,
		clientSecret: clientSecret,
	}

	// todo: validate valid values here
	return &a, nil
}

func basicAuth(a *ClientCredentialAuth) string {
	return b64.StdEncoding.EncodeToString([]byte(a.clientId + ":" + a.clientSecret))
}

func (a *ClientCredentialAuth) Authenticate() (*AuthResponse, error) {
	q := url.Values{}
	q.Add("grant_type", "client_credentials")
	u := url.URL{}

	u.Scheme = "https"
	u.Host = a.host
	u.Path = "/services/oauth2/token"

	requestURL := u.String()

	req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(q.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+basicAuth(a))

	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ar := AuthResponse{}
	err = json.Unmarshal(b, &ar)
	return &ar, err
}
