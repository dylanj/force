package force

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/net/publicsuffix"
)

type connectionRequest struct {
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

type connectionResponse struct {
	ClientID string `json:"clientId"`
	Advice   *struct {
		Interval  int    `json:"interval"`
		Timeout   int    `json:"timeout"`
		Reconnect string `json:"reconnect"`
	} `json:"advice"`
	Successful bool         `json:"successful"`
	Data       *DataMessage `json:"data"`
}

type DataMessage struct {
	Schema  string           `json:"schema"`
	Payload *json.RawMessage `json:"payload"`
	Event   struct {
		ReplayId int `json:"replayId"`
	} `json:"event"`
}

type handshakeResponse struct {
	Ext struct {
		Replay        bool `json:"replay"`
		PayloadFormat bool `json:"payload.format"`
	} `json:"ext"`
	MinimumVersion           string   `json:"minimumVersion"`
	ClientID                 string   `json:"clientId"`
	SupportedConnectionTypes []string `json:"supportedConnectionTypes"`
	Channel                  string   `json:"channel"`
	Version                  string   `json:"version"`
	Successful               bool     `json:"successful"`
}

type StreamingMessage struct {
	Payload  json.RawMessage
	ReplayId uint
}

type StreamingClient struct {
	messages chan (StreamingMessage)

	channel    string
	sf         *Client
	httpClient *http.Client
	clientId   string
	replayId   int
}

func (c *StreamingClient) handshake() (*handshakeResponse, error) {
	msg := connectionRequest{
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
	// hack, we store version as vXX.X for some reason.
	// todo: fix this.
	cometdpath := "/cometd/" + strings.Replace(c.sf.version, "v", "", 1)

	return c.sf.postWithClient(c.httpClient, cometdpath, payload)
}

func (c *StreamingClient) connect() ([]*connectionResponse, error) {
	connectMessage := connectionRequest{
		Channel:        "/meta/connect",
		ClientID:       c.clientId,
		ConnectionType: "long-polling",
	}

	b, err := c.post(connectMessage)
	if err != nil {
		return []*connectionResponse{}, err
	}

	r := []*connectionResponse{}
	err = json.Unmarshal(b, &r)
	if err != nil {
		return []*connectionResponse{}, err
	}

	return r, nil
}

func (c *StreamingClient) subscribe(channel string, replayId int) error {
	replayExt := make(map[string]int)
	replayExt[channel] = replayId

	subscribeRequest := connectionRequest{
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

	r := []*connectionResponse{}
	err = json.Unmarshal(b, &r)
	if err != nil {
		return err
	}

	if r[0].Successful == false {
		return errors.New("invalid subscription")
	}

	return nil
}

func newStreamingClient(sf *Client, channel string, replayId int) StreamingClient {
	cookiejarOptions := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, _ := cookiejar.New(&cookiejarOptions)

	httpClient := http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
	}

	fmt.Println("response headers")
	return StreamingClient{
		channel:    channel,
		replayId:   replayId,
		messages:   make(chan StreamingMessage),
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
	// todo: grab advice from connection. timeouts

	err = c.subscribe(c.channel, c.replayId)
	if err != nil {
		return err
	}
	return nil
}

func (c *StreamingClient) poll(handler func(m *DataMessage) error) {
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

func (sfc *Client) Subscribe(channel string, replayId int, handler func(m *DataMessage) error) {
	c := newStreamingClient(sfc, channel, replayId)
	err := c.begin() // handshake, connect, subscribe
	if err != nil {
		return
	}
	c.poll(handler)
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
