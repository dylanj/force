package force

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type UserPassAuth struct {
	host          string
	username      string
	password      string
	securityToken string
	clientId      string
	clientSecret  string
}

func AuthUserPass(host, username, password, token, clientId, clientSecret string) (*UserPassAuth, error) {
	a := UserPassAuth{
		host:          host,
		username:      username,
		password:      password,
		securityToken: token,
		clientId:      clientId,
		clientSecret:  clientSecret,
	}

	// todo: validate valid values here
	return &a, nil
}

func (a *UserPassAuth) Authenticate() (*AuthResponse, error) {
	q := url.Values{}
	q.Add("grant_type", "password")
	q.Add("client_id", a.clientId)
	q.Add("client_secret", a.clientSecret)
	q.Add("username", a.username)
	q.Add("password", a.password+a.securityToken)

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
