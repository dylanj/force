package force

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

type streamingChannel struct {
}

/*
{
"data": {
"schema": "dffQ2QLzDNHqwB8_sHMxdA",
"payload": {
"CreatedDate": "2017-04-09T18:31:40.517Z",
"CreatedById": "005D0000001cSZs",
"Printer_Model__c": "XZO-5",
"Serial_Number__c": "12345",
"Ink_Percentage__c": 0.2
},
"event": {
"replayId": 2
}
},
"channel": "/event/Low_Ink__e"
}
*/

//subscribeParams := `{ "channel": "/meta/subscribe", "clientID": "` + forceAPI.stream.ClientID + `", "subscription": "` + eventString + `"}`

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
	Advice   struct {
		Interval  int    `json:"interval"`
		Timeout   int    `json:"timeout"`
		Reconnect string `json:"reconnect"`
	} `json:"advice"`
	Successful bool `json:"successful"`
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
	Payload  []byte
	ReplayId uint
	Channel  string
}

type StreamingClient struct {
	handler  func(StreamingMessage) error
	messages chan (StreamingMessage)

	sf       *Client
	clientId string
	replayId int
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

func (c *StreamingClient) Poll() {
	for {
		//r, err := c.connect()
	}
}

func (c *StreamingClient) post(payload any) ([]byte, error) {
	// hack, we store version as vXX.X for some reason.
	// todo: fix this.
	cometdpath := "/cometd/" + strings.Replace(c.sf.version, "v", "", 1)

	//c.sf.auth

	return c.sf.Post(cometdpath, payload)
}

func (c *StreamingClient) connect() (*connectionResponse, error) {
	connectMessage := connectionRequest{
		Channel:        "/meta/connect",
		ClientID:       c.clientId,
		ConnectionType: "long-polling",
	}

	b, err := c.post(connectMessage)
	if err != nil {
		return nil, err
	}

	fmt.Println("connect")
	fmt.Println(string(b))

	r := []connectionResponse{}
	err = json.Unmarshal(b, &r)
	if err != nil {
		return nil, err
	}

	return &r[0], nil
}

func (c *StreamingClient) TestConnect() {

	h, err := c.handshake()
	spew.Dump(h)
	if err != nil {
		spew.Dump(err)
	}

	c.clientId = h.ClientID
	fmt.Println("got client id", c.clientId)

	cr, err := c.connect()
	spew.Dump(cr)

	//		subscribeParams := `{ "channel": "/meta/subscribe", "clientID": "` + forceAPI.stream.ClientID + `", "subscription": "` + topicString + `"}`

	replayExt := make(map[string]int)
	ch := "/event/S5_Sync__e"
	replayExt[ch] = 17799646

	subscribeRequest := connectionRequest{
		Channel:      "/meta/subscribe",
		ClientID:     c.clientId,
		Subscription: "/event/S5_Sync__e",
		Ext: extensions{
			Replay: replayExt,
		},
	}

	r, err := c.post(subscribeRequest)
	spew.Dump(err)
	fmt.Println(string(r))
	cr, err = c.connect()
	spew.Dump(cr)

}

func NewStreamingClient(sf *Client, replayId int, handler func(StreamingMessage) error) StreamingClient {
	fmt.Println("response headers")
	return StreamingClient{
		handler:  handler,
		replayId: replayId,
		messages: make(chan StreamingMessage),
		sf:       sf,
	}
}

/*
func (c *StreamingClient) Subscribe(channel string, replayId int, handler func(m StreamingMessage) error) {
	c.handlers[channel] = handler
}

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
