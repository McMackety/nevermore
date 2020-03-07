package main

import (
	"bufio"
	"fmt"
	"github.com/McMackety/nevermore/config"
	"github.com/McMackety/nevermore/database"
	"github.com/McMackety/nevermore/field"
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
	field.CreateField()
	field.CurrentField.Run()
	database.InitDatabase()

	// CLI app down here, mostly used for pre-gui debugging

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		parts := strings.Split(text, " ")
		switch parts[0] {
		case "enableAll":
			field.CurrentField.EnableAllRobots()
			continue
		case "disableAll":
			field.CurrentField.DisableAllRobots()
			continue
		case "startMatch":
			err := field.CurrentField.StartField()
			if err != nil {
				log.Println(err.Error())
			}
			continue
		case "stopMatch":
			err := field.CurrentField.StopField(true)
			if err != nil {
				log.Println(err.Error())
			}
			continue
		case "startTest":
			field.CurrentField.MatchLevel = field.MATCHTEST
			continue
		case "stopTest":
			field.CurrentField.MatchLevel = field.PRACTICE
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
		case "removeTeamByStation":
			if station, err := strconv.Atoi(parts[1]); err == nil {
				if teamNum, ok := field.CurrentField.AllianceStationToTeam[field.AllianceStation(station)]; ok {
					if driverStation := field.CurrentField.GetDriverStationByTeamNum(teamNum); driverStation != nil {
						delete(field.CurrentField.AllianceStationToTeam, field.AllianceStation(station))
						driverStation.Kick()
					}
				}
			}
			println("Improper usage of enable: Usage: removeTeamByStation <station>")
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
