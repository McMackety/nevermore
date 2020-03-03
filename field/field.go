package field

import (
	"errors"
	"fmt"
	"github.com/McMackety/nevermore/web"
	"log"
	"net"
	"strings"
	"time"
)

var correctNetwork bool = false
var CurrentField Field

type Field struct {
	TeamNumberToDriverStation map[int]DriverStation `json:"teamNumberToDriverStation"`
	AllianceStationToTeam map[AllianceStation]int `json:"allianceStationToTeam"`
	MatchPaused bool `json:"matchPaused"`
	MatchStarted bool `json:"matchStarted"`
	MatchReady bool `json:"matchReady"`
	MatchDone bool `json:"matchDone"`
	MatchStartedAt time.Time `json:"matchStartedAt"`
	MatchLevel Level `json:"matchLevel"`
	MatchNumber int `json:"matchNumber"`
	TimeLeft int `json:"timeLeft"`
	UDPSocket *net.UDPConn `json:"-"`
	EventName string `json:"eventName"`
	CurrentPhase string `json:"currentPhase"`
}

// InitField starts the networking for the FMS
func InitField() {
	log.Println("Initializing Field...")
	// Check for a network interface with 10.0.100.5
	addrs, _ := net.InterfaceAddrs()
	log.Println("Scanning Network Interfaces: ")
	for _, addr := range addrs {
		log.Println("Found IP: " + addr.String())
		if strings.HasPrefix(addr.String(), "10.0.100.5") {
		 	correctNetwork = true
		}
	}
	if !correctNetwork {
		log.Println("Couldn't find network interface with an IP of 10.0.100.5, the FMS will not be available however you can still use the Web Interface!")
	}

	CurrentField = Field{
		TeamNumberToDriverStation: make(map[int]DriverStation),
		AllianceStationToTeam: make(map[AllianceStation]int),
		MatchPaused: false,
		MatchStarted: false,
		MatchReady: false,
		MatchDone: false,
		MatchStartedAt: time.Now(),
		MatchLevel: MATCHTESTLEVEL,
		MatchNumber: 0,
		EventName: "EAO",
	}

	go CurrentField.fieldTimer()
	go CurrentField.tick()
	go CurrentField.listenTCP()
	go CurrentField.listenUDP()
}

func (field *Field) GetDriverStationByTeamNum(teamNum int) *DriverStation {
	if val, ok := field.TeamNumberToDriverStation[teamNum]; ok {
		return &val
	}
	return nil
}


func (field *Field) GetDriverStationByIP(ip net.Addr) *DriverStation {
	for _, driverStation := range field.TeamNumberToDriverStation {
		if driverStation.TCPSocket.RemoteAddr() == ip {
			return &driverStation
		}
	}
	return nil
}

func (field *Field) GetAllianceStationFromTeamNum(teamNum int) AllianceStation {
	for allianceStation, team := range field.AllianceStationToTeam {
		if team == teamNum {
			return allianceStation
		}
	}
	return 0 // Returns RED1
}

func (field *Field) IsTeamInMatch(teamNum int) bool {
	for _, team := range field.AllianceStationToTeam {
		if team == teamNum {
			return true
		}
	}
	return false
}

func (field *Field) DisableAllRobots() {
	field.MatchPaused = true
	for _, driverStation := range field.TeamNumberToDriverStation {
		driverStation.Enabled = false
	}
}

func (field *Field) DisableAllRobotsNoPause() {
	for _, driverStation := range field.TeamNumberToDriverStation {
		driverStation.Enabled = false
	}
}

func (field *Field) SetAllRobotModes(mode Mode) {
	for _, driverStation := range field.TeamNumberToDriverStation {
		driverStation.Mode = mode
	}
}

func (field *Field) EnableAllRobots() {
	field.MatchPaused = false
	for _, teamNum := range field.AllianceStationToTeam {
		driverStation, ok := field.TeamNumberToDriverStation[teamNum]
		if !ok {
			continue
		}
		driverStation.Enabled = true
	}
}

func (field *Field) KickAllDriverStations() {
	for _, driverStation := range field.TeamNumberToDriverStation {
		driverStation.Kick()
	}
}

func (field *Field) SetupField(matchNum int, tournamentLevel Level, red1 int, red2 int, red3 int, blue1 int, blue2 int, blue3 int) {
	field.KickAllDriverStations()
	field.MatchStarted = false
	field.MatchDone = false
	field.MatchReady = false
	field.MatchLevel = tournamentLevel
	field.AllianceStationToTeam[REDSTATION1] = red1
	field.AllianceStationToTeam[REDSTATION2] = red2
	field.AllianceStationToTeam[REDSTATION3] = red3
	field.AllianceStationToTeam[BLUESTATION1] = blue1
	field.AllianceStationToTeam[BLUESTATION2] = blue2
	field.AllianceStationToTeam[BLUESTATION3] = blue3
	web.WebEventsServer.EmitJSONAll("matchStatus", field)
}

func (field *Field) StartField() error {
	if !field.MatchReady {
		return errors.New("the match is not ready, not all driverStations have connected")
	}
	if field.MatchStarted {
		return errors.New("the match has already started, setup the match before you restart it")
	}

	web.WebEventsServer.EmitJSONAll("matchStatus", field)
	field.MatchStarted = true
	return nil
}

func (field *Field) fieldTimer() {
	if field.MatchDone {
		return
	}
	if field.MatchStarted && !field.MatchPaused {
		field.TimeLeft--
		web.WebEventsServer.EmitJSONAll("newTime", field.TimeLeft)
		if field.TimeLeft > TransitionLength + TeleopLength + EndgameLength {
			if field.CurrentPhase != "auto" {
				field.SetAllRobotModes(AUTONOMOUSMODE)
				web.WebEventsServer.EmitJSONAll("changePhase", "auto")
				field.EnableAllRobots()
			}
			field.CurrentPhase = "auto"
		} else if field.TimeLeft > TeleopLength + EndgameLength {
			if field.CurrentPhase != "transition" {
				web.WebEventsServer.EmitJSONAll("changePhase", "transition")
				field.DisableAllRobotsNoPause()
			}
			field.CurrentPhase = "transition"
		} else if field.TimeLeft > EndgameLength {
			if field.CurrentPhase != "teleop" {
				field.SetAllRobotModes(TELEOPMODE)
				web.WebEventsServer.EmitJSONAll("changePhase", "teleop")
				field.EnableAllRobots()
			}
			field.CurrentPhase = "teleop"
		} else if field.TimeLeft > 0 {
			if field.CurrentPhase != "endgame" {
				web.WebEventsServer.EmitJSONAll("changePhase", "endgame")
			}
			field.CurrentPhase = "endgame"
		} else {
			field.DisableAllRobotsNoPause()
			field.MatchDone = true
			web.WebEventsServer.EmitJSONAll("matchDone", true)
		}

	}
	time.Sleep(time.Second)
	field.fieldTimer()
}

func (field *Field) tick() {
	// Check if all teams are on field in order to say that the game is ready.
	hasAllTeamsOnField := false
	for _, teamNum := range field.AllianceStationToTeam {
		teamIsOnField := false
		for _, driverStation := range field.TeamNumberToDriverStation {
			if driverStation.TeamNumber == teamNum {
				teamIsOnField = true
			}
		}
		if teamIsOnField {
			hasAllTeamsOnField = true
		} else {
			hasAllTeamsOnField = false
			break
		}
	}
	if hasAllTeamsOnField {
		field.MatchReady = true
		web.WebEventsServer.EmitJSONAll("matchReady", true)
	} else {
		field.MatchReady = false
		web.WebEventsServer.EmitJSONAll("matchReady", false)
	}
	for _, driverStation := range field.TeamNumberToDriverStation {
		driverStation.tick()
	}
	time.Sleep(time.Millisecond * 500)
	field.tick()
}

func (field *Field) createDriverStation(teamNum int, socket net.Conn, udpSocket net.Conn) {
	driverStation := DriverStation{
		TCPSocket:        socket,
		CurrentField:     field,
		TeamNumber:       teamNum,
		EmergencyStopped: false,
		Comms:            false,
		RadioPing:        false,
		RioPing:          false,
		Enabled:          false,
		Mode:             TESTMODE,
		BatteryVoltage:   0.0,
		UDPSequenceNum:   0,
		LostPacketsNum:   0,
		AverageTripTime:  0,
		LastUDPMessage:   time.Now(),
		Station:          field.GetAllianceStationFromTeamNum(teamNum),
		UDPConn: 		  udpSocket,
	}

	if newStation, ok := field.TeamNumberToDriverStation[teamNum]; ok {
		driverStation = newStation
		return
	}
	field.TeamNumberToDriverStation[teamNum] = driverStation

	if !field.IsTeamInMatch(teamNum) {
		driverStation.Status = WAITINGSTATUS
	} else {
		driverStation.Status = GOODSTATUS
	}

	driverStation.sendStationInfo()
	driverStation.sendEventName()
}

func (field *Field) listenTCP() {
	listener, err := net.Listen("tcp", ":1750")
	if err != nil {
		log.Println("Couldn't start FMS TCP Server: " + err.Error())
		return
	}
	log.Println("The FMS started a TCP Server on 10.0.100.5:1750!")
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go func() {
			// errorNum is used to detect if the tcpConn had closed.
			errorNum := 0
			for {
				if errorNum > 3 {
					conn.Close()
					return
				}
				var bytes [5]byte
				_, err := conn.Read(bytes[:])
				if err != nil {
					errorNum++
					continue
				}
				if bytes[2] == 0x18 {
					teamNum := (int(bytes[3]) << 8) + int(bytes[4])
					ipAddress, _, err := net.SplitHostPort(conn.RemoteAddr().String())
					if err != nil {
						continue
					}
					udpConn, err := net.Dial("udp4", fmt.Sprintf("%s:%d", ipAddress, 1121))
					field.createDriverStation(teamNum, conn, udpConn)
				}
			}
		}()
	}
}

func (field *Field) listenUDP() {
	udpAddr, err := net.ResolveUDPAddr("udp4", ":1160")
	if err != nil {
		log.Println("Bad Address for UDP: " + err.Error())
		return
	}

	listener, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		log.Println("Couldn't start FMS UDP Server: " + err.Error())
		return
	}

	log.Println("The FMS started a UDP Server on 10.0.100.5:1160!")

	field.UDPSocket = listener

	defer listener.Close()

	var bytes [50]byte
	for {
		listener.Read(bytes[:])
		eStopped := (int(bytes[3]) >> 7 & 0x01) == 1
		comms := (int(bytes[3]) >> 5 & 0x01) == 1
		radioPing := (int(bytes[3]) >> 4 & 0x01) == 1
		rioPing := (int(bytes[3]) >> 3 & 0x01) == 1
		enabled := (int(bytes[3]) >> 2 & 0x01) == 1
		mode := Mode(int(bytes[3]) & 0x03)
		teamNum := (int(bytes[4]) << 8) + int(bytes[5])
		batteryVoltage := float64(bytes[6]) + float64(bytes[7])/256

		if driverStation := field.GetDriverStationByTeamNum(teamNum); driverStation != nil {
			driverStation.receiveUDP(eStopped, comms, radioPing, rioPing, enabled, mode, batteryVoltage)
		}
	}
}