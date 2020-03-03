package main

import (
	"bufio"
	"fmt"
	"github.com/McMackety/nevermore/config"
	"github.com/McMackety/nevermore/field"
	"github.com/McMackety/nevermore/web"
	"log"
	"os"
	"strconv"
	"strings"
)

var Version string = "0.0.1" // This will be injected at build time, don't worry about it :)
var GitCommit string = "dev" // This will be injected at build time, don't worry about it :)

func main() {
	log.Printf("Starting nevermore v%s (Commit %s)", Version, GitCommit)
	config.LoadConfig()
	web.StartWebServer()
	field.InitField()
	log.Printf("Nevermore v%s has started.", Version)

	// CLI app down here, mostly used for pre-gui debugging

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		parts := strings.Split(text, " ")
		switch parts[0] {
		case "enable":
			if out, err := strconv.Atoi(parts[1]); err == nil {
				if driverStation, ok := field.CurrentField.TeamNumberToDriverStation[out]; ok {
					if enabled, err := strconv.ParseBool(parts[2]); err == nil {
						driverStation.Enabled= enabled
						continue
					}
				}
			}
			println("Improper usage of enable: Usage: enable <teamNum> <true|false>")
			continue
		case "addTeam":
			if station, err := strconv.Atoi(parts[1]); err == nil {
				if team, err := strconv.Atoi(parts[2]); err == nil {
					field.CurrentField.AllianceStationToTeam[field.AllianceStation(station)] = team
					continue
				}
			}
			println("Improper usage of enable: Usage: addTeam <station> <teamNum>")
			continue
		case "station":
			if out, err := strconv.Atoi(parts[1]); err == nil {
				if driverStation, ok := field.CurrentField.TeamNumberToDriverStation[out]; ok {
					if station, err := strconv.Atoi(parts[2]); err == nil {
						driverStation.Station = field.AllianceStation(station)
					}
				}
			}
			continue
		}
	}
}