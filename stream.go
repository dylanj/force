package force

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/net/publicsuffix"
)

type StreamMessage struct {
	Schema  string           `json:"schema"`
	Payload *json.RawMessage `json:"payload"`
	Event   struct {
		ReplayId int `json:"replayId"`
	} `json:"event"`
}

type StreamingClient struct {
	channel    string
	sf         *Client
	httpClient *http.Client
	clientId   string
	replayId   int
}

type streamRequest struct {
	Channel                  string   `json:"channel"`
	Subscription             string   `json:"subscription,omitempty"`
	ConnectionType           string   `json:"connectionType,omitempty"`
	Version                  string   `json:"version,omitempty"`
	ClientID                 string   `json:"clientId"`
	SupportedConnectionTypes []string `json:"supportedConnectionTypes,omitempty"`

	Ext extensions `json:"ext,omitempty"`
}

type extensions struct {
	Replay map[string]int `json:"replay,omitempty"`
}

type streamResponse struct {
	ClientID string `json:"clientId"`
	Advice   *struct {
		Interval  int    `json:"interval"`
		Timeout   int    `json:"timeout"`
		Reconnect string `json:"reconnect"`
	} `json:"advice"`
	Data       *StreamMessage `json:"data"`
	Error      string         `json:"error"`
	Successful bool           `json:"successful"`
}

type handshakeResponse struct {
	Ext struct {
		Replay        bool `json:"replay"`
		PayloadFormat bool `json:"payload.format"`
	} `json:"ext"`
	Version                  string   `json:"version"`
	MinimumVersion           string   `json:"minimumVersion"`
	ClientID                 string   `json:"clientId"`
	SupportedConnectionTypes []string `json:"supportedConnectionTypes"`
	Channel                  string   `json:"channel"`
	Successful               bool     `json:"successful"`
}

// todo: move this to client.go
func (sfc *Client) Subscribe(channel string, replayId int, handler func(m *StreamMessage) error) error {
	c := newStreamingClient(sfc, channel, replayId)
	err := c.begin() // handshake, connect, subscribe
	if err != nil {
		return err
	}
	c.poll(handler)
	return nil
}

// todo: create go thread to handle messages received
func (c *StreamingClient) poll(handler func(m *StreamMessage) error) {
	for {
		cr, err := c.connect()
		if err != nil {
			spew.Dump(err)
			continue
		}
		for _, m := range cr {
			if m.Data != nil {
				handler(m.Data)
			} else {
				fmt.Println("whats going on here")
				spew.Dump(m)
			}
		}
	}
	fmt.Println("done")
}

func (c *StreamingClient) handshake() (*handshakeResponse, error) {
	msg := streamRequest{
		Channel:                  "/meta/handshake",
		SupportedConnectionTypes: []string{"long-polling"},
		Version:                  "1.0",
	}

	r, err := c.post(msg)
	if err != nil {
		return nil, err
	}

	h := []handshakeResponse{}
	err = json.Unmarshal(r, &h)
	if err != nil {
		return nil, err
	}

	return &h[0], nil
}

func (c *StreamingClient) post(payload any) ([]byte, error) {
	return c.sf.postWithClient(c.httpClient, "/cometd/"+c.sf.version, payload)
}

func (c *StreamingClient) connect() ([]*streamResponse, error) {
	connectMessage := streamRequest{
		Channel:        "/meta/connect",
		ClientID:       c.clientId,
		ConnectionType: "long-polling",
	}

	b, err := c.post(connectMessage)
	if err != nil {
		return []*streamResponse{}, err
	}

	r := []*streamResponse{}
	err = json.Unmarshal(b, &r)
	if err != nil {
		return []*streamResponse{}, err
	}

	return r, nil
}

func (c *StreamingClient) subscribe(channel string, replayId int) error {
	replayExt := make(map[string]int)
	replayExt[channel] = replayId

	subscribeRequest := streamRequest{
		Channel:      "/meta/subscribe",
		ClientID:     c.clientId,
		Subscription: channel,
		Ext: extensions{
			Replay: replayExt,
		},
	}

	b, err := c.post(subscribeRequest)
	if err != nil {
		return err
	}

	r := []*streamResponse{}
	err = json.Unmarshal(b, &r)
	if err != nil {
		return err
	}

	if r[0].Successful == false {
		return errors.New(r[0].Error)
	}

	return nil
}

func newStreamingClient(sf *Client, channel string, replayId int) StreamingClient {
	// todo: make a copy of the force client, but first refactor the auth so we
	// can reauth and update all clients
	cookiejarOptions := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, _ := cookiejar.New(&cookiejarOptions)

	httpClient := http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
	}

	return StreamingClient{
		channel:    channel,
		replayId:   replayId,
		httpClient: &httpClient,
		sf:         sf,
	}
}

func (c *StreamingClient) begin() error {
	h, err := c.handshake()
	if err != nil {
		return err
	}

	c.clientId = h.ClientID

	cr, err := c.connect()
	if err != nil {
		return err
	}

	timeoutVal := cr[0].Advice.Timeout
	timeoutDur := time.Duration(timeoutVal) * time.Millisecond

	c.httpClient.Timeout = time.Duration(timeoutDur)

	err = c.subscribe(c.channel, c.replayId)
	if err != nil {
		return err
	}
	return nil
}

/*
func (c *StreamingClient) Start() {
	//c.connect()
	for k, v := range c.handlers {
		//
		fmt.Println(k, v)
	}

	for {
		select {
		case m := <-c.messages:
			f := c.handlers[m.Channel]
			f(m)
		}
	}
}

func (c *StreamingClient) consume(m StreamingMessage) {
	c.messages <- m
}
*/
