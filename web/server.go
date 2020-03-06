package web

import (
	"github.com/McMackety/nevermore/config"
	"github.com/McMackety/nevermore/webevents"
	"log"
	"net/http"
)

func StartWebServer() {
	log.Println("WebEvents Server listening on " + config.DefaultConfig.WebSocketListenAddress)
	webevents.WebEventsServer = webevents.CreateServer()
	webevents.WebEventsServer.OnClientJoin(func(client *webevents.Client) {
		client.Set("loginAttempts", 0)
		client.OnEvent("login", login)
		client.OnEvent("getAllUsers", getAllUsers)
		client.OnEvent("getUserByID", getUserByID)
		client.OnEvent("getUserByUsername", getUserByUsername)
		client.OnEvent("createUser", createUser)
		client.OnEvent("getMatch", getMatch)
		client.OnEvent("setupMatch", setupMatch)
		client.OnEvent("disableAll", disableAll)
		client.OnEvent("enableAll", enableAll)

	})
	go http.ListenAndServe(config.DefaultConfig.WebSocketListenAddress, nil)
}
