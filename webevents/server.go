package webevents

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"go/types"
	"net/http"
)

var invalidMessage = MessageWrapper{"error", Error{"invalidMessage", "That message as invalid!"}}

var unknownEvent = MessageWrapper{"error", Error{"unknownEvent", "That event doesn't exist!"}}

type Server struct {
	joinHandlers []func(client *Client)
	closeHandlers []func(client *Client)
	Clients map[*websocket.Conn]Client
	upgrader websocket.Upgrader
}

type Error struct {
	Name string `json:"name"`
	Message string `json:"message"`
}

type MessageWrapper struct {
	EventName string `json:"eN"`
	Message interface{} `json:"m"`
}

type MessageDecoder struct {
	EventName string `json:"eN"`
	Message json.RawMessage `json:"m"`
}

type SubscriptionMessage struct {
	EventToSubscribeTo string `json:"eventName"`
}

// CreateServer creates a WebEvents server, IT DOES NOT LISTEN TO A ADDRESS YOU NEED TO ADD THAT YOURSELF
func CreateServer() (server *Server) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	server = new(Server)
	server.upgrader = upgrader
	http.HandleFunc("/ws", server.handleWs)
	return server
}

func (server *Server) EmitJSONAll(event string, json interface{}) {
	for _, client := range server.Clients {
		client.EmitJSON(event, json)
	}
}

func (server *Server) handleWs(w http.ResponseWriter, r *http.Request) {
	ws, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(400)
	}

	client := Client{
		Websocket: ws,
		handlers: make(map[string]func(client *Client, json json.RawMessage)),
		StoredValues: make(map[interface{}]interface{}),
	}

	server.Clients[ws] = client

	for _, callback := range server.joinHandlers {
		callback(&client)
	}

	go server.handleClient(&client)
}

func (server *Server) handleClient(client *Client) {
	for {
		var jsonMessage MessageDecoder
		err := client.Websocket.ReadJSON(&jsonMessage)

		// Error reading message.
		if err != nil {
			err = client.Websocket.WriteJSON(invalidMessage)
			// Error writing message means disconnect them.
			if err != nil {
				_ = client.Websocket.Close()
				break
			}
			continue
		}

		if jsonMessage.EventName == "subscribe" {
			var message SubscriptionMessage
			err := json.Unmarshal(jsonMessage.Message, &message)
			if err != nil {
				client.Websocket.WriteJSON(invalidMessage)
				continue
			}
			client.Subscriptions[message.EventToSubscribeTo] = types.Nil{}
			continue
		}
		if jsonMessage.EventName == "unsubscribe" {
			var message SubscriptionMessage
			err := json.Unmarshal(jsonMessage.Message, &message)
			if err != nil {
				client.Websocket.WriteJSON(invalidMessage)
				continue
			}
			delete(client.Subscriptions, message.EventToSubscribeTo)
			continue
		}
		if callback, ok := client.handlers[jsonMessage.EventName]; ok {
			callback(client, jsonMessage.Message)
		} else {
			err = client.Websocket.WriteJSON(unknownEvent)
			// Error writing message means disconnect them.
			if err != nil {
				_ = client.Websocket.Close()
				break
			}
		}
	}
	delete(server.Clients, client.Websocket)
	for _, callback := range server.closeHandlers {
		callback(client)
	}
}

// OnClientJoin adds a client joined callback
func (server *Server) OnClientJoin(callback func(client *Client)) {
	server.joinHandlers = append(server.joinHandlers, callback)
}

// OnClientClose adds a client closed callback
func (server *Server) OnClientClose(callback func(client *Client)) {
	server.closeHandlers = append(server.closeHandlers, callback)
}