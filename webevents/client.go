package webevents

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"go/types"
)

type Client struct {
	handlers map[string]func(client *Client, json json.RawMessage)
	Websocket *websocket.Conn
	StoredValues map[interface{}]interface{}
	Subscriptions map[string]types.Nil // Maps are faster to lookup in, thus just use a nil type.
}

func (client *Client) OnEvent(event string, callback func(client *Client, json json.RawMessage)) {
	client.handlers[event] = callback
}

func (client *Client) EmitJSON(event string, json interface{}) {
	if _, ok := client.Subscriptions[event]; !ok { // The Client hasn't subscribed to this event, thus it is useless.
		return
	}
	message := MessageWrapper{
		EventName: event,
		Message:   json,
	}
	err := client.Websocket.WriteJSON(message)

	// Error writing message means disconnect them.
	if err != nil {
		client.Websocket.Close()
	}
}

func (client *Client) EmitData(event string, data []byte) {
	if _, ok := client.Subscriptions[event]; !ok { // The Client hasn't subscribed to this event, thus it is useless.
		return
	}
	message := MessageWrapper{
		EventName: event,
		Message:   data,
	}
	err := client.Websocket.WriteJSON(message)

	// Error writing message means disconnect them.
	if err != nil {
		client.Websocket.Close()
	}
}

func (client *Client) Close() {
	client.Websocket.Close()
}

func (client *Client) Get(key interface{}) interface{} {
	return client.StoredValues[key]
}

func (client *Client) Set(key interface{}, value interface{}) {
	client.StoredValues[key] = value
}