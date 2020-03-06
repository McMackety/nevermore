package web

import (
	"encoding/json"
	"github.com/McMackety/nevermore/field"
	"github.com/McMackety/nevermore/webevents"
)

func getMatch(client *webevents.Client, data json.RawMessage) {
	if client.Get("loggedIn") != true {
		client.EmitJSON("notAuthenticated", true)
		return
	}
	client.EmitJSON("getMatchReply", field.CurrentField)
}

type SetupMatchMessage struct {
	MatchNumber     int         `json:"matchNum"`
	TournamentLevel field.Level `json:"tournamentLevel"`
	RedStation1     int         `json:"red1"`
	RedStation2     int         `json:"red2"`
	RedStation3     int         `json:"red3"`
	BlueStation1    int         `json:"blue1"`
	BlueStation2    int         `json:"blue2"`
	BlueStation3    int         `json:"blue3"`
}

func setupMatch(client *webevents.Client, data json.RawMessage) {
	if client.Get("loggedIn") != true {
		client.EmitJSON("notAuthenticated", true)
		return
	}
	var message SetupMatchMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		client.EmitJSON("error", "Invalid Message! "+err.Error())
		return
	}
	field.CurrentField.SetupField(message.MatchNumber, message.TournamentLevel, message.RedStation1, message.RedStation2, message.RedStation3, message.BlueStation1, message.BlueStation2, message.BlueStation3)
}

func startMatch(client *webevents.Client, data json.RawMessage) {
	if client.Get("loggedIn") != true {
		client.EmitJSON("notAuthenticated", true)
		return
	}
	var message SetupMatchMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		client.EmitJSON("error", "Invalid Message! "+err.Error())
		return
	}
	err = field.CurrentField.StartField()
	if err != nil {
		client.EmitJSON("error", err.Error())
	}
}

func disableAll(client *webevents.Client, data json.RawMessage) {
	if client.Get("loggedIn") != true {
		client.EmitJSON("notAuthenticated", true)
		return
	}
	field.CurrentField.DisableAllRobots()
}

func enableAll(client *webevents.Client, data json.RawMessage) {
	if client.Get("loggedIn") != true {
		client.EmitJSON("notAuthenticated", true)
		return
	}
	field.CurrentField.EnableAllRobots()
}
