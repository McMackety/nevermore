package web

import (
	"github.com/McMackety/nevermore/config"
	"github.com/McMackety/nevermore/webevents"
	"log"
	"net/http"
)

var WebEventsServer *webevents.Server

func StartWebServer() {
	log.Println("WebEvents Server listening on " + config.DefaultConfig.WebSocketListenAddress)
	WebEventsServer = webevents.CreateServer()
	go http.ListenAndServe(config.DefaultConfig.WebSocketListenAddress, nil)
}