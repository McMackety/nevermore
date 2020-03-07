package field

import (
	"errors"
	"fmt"
	"github.com/McMackety/nevermore/scoring"
	"log"
	"net"
	"strings"
	"time"
)

// The current field, should be obvious
var CurrentField *Field

// Field is the structure the represents a FRC field
type Field struct {
	MatchNumber               int `json:"matchNum"`
	MatchState				  State `json:"matchState"`
	TimeLeft                  int `json:"timeLeft"`
	EventName                 string `json:"eventName"`
	CurrentPhase              Phase `json:"currentPhase"`
	MatchLevel                Level `json:"matchLevel"`
	MatchStartedAt            time.Time `json:"matchStartedAt"`
	Scorer 					  scoring.ScoringInterface
	TeamNumberToDriverStation map[int]*DriverStation `json:"teamNumberToDriverStation"`
	AllianceStationToTeam     map[AllianceStation]int `json:"allianceStationToTeam"`
	UDPSocket                 *net.UDPConn `json:"-"`
	Log 					  []string `json:"-"`
}

// CreateField creates a field
func CreateField() {
	log.Println("Initializing Field...")
	// Check for a network interface with 10.0.100.5
	correctNetwork := false
	addrs, _ := net.InterfaceAddrs()
	log.Println("Scanning Network Interfaces...")
	for _, addr := range addrs {
		if strings.HasPrefix(addr.String(), "10.0.100.5") {
			correctNetwork = true
		}
	}
	if !correctNetwork {
		log.Panicln("Couldn't find network interface with an IP of 10.0.100.5, the FMS can't startup!")
	}
	log.Println("Found network interface with 10.0.100.5!")

	field := Field{
		TeamNumberToDriverStation: make(map[int]*DriverStation),
		AllianceStationToTeam:     make(map[AllianceStation]int),
		MatchState:				   NOTREADY,
		Scorer: 				   scoring.CreateScoringInterface(),
		MatchStartedAt:            time.Now(),
		MatchLevel:                MATCHTEST,
		MatchNumber:               0,
		EventName:                 "EAO",
		CurrentPhase: 			   NOTHING,
	}
	CurrentField = &field
}

// Starts the FMS's networking
func (field *Field) Run() {
	go field.fieldTimer()
	go field.tick()
	go field.listenTCP()
	go field.listenUDP()
}

// Sets up the field from scratch
func (field *Field) SetupField(matchNum int, tournamentLevel Level, red1 int, red2 int, red3 int, blue1 int, blue2 int, blue3 int) {
	field.KickAllDriverStations()
	field.MatchNumber = matchNum
	field.MatchState = NOTREADY
	field.MatchLevel = tournamentLevel
	field.AllianceStationToTeam[RED1] = red1
	field.AllianceStationToTeam[RED2] = red2
	field.AllianceStationToTeam[RED3] = red3
	field.AllianceStationToTeam[BLUE1] = blue1
	field.AllianceStationToTeam[BLUE2] = blue2
	field.AllianceStationToTeam[BLUE3] = blue3
	field.Log = make([]string, 50)
}

// Starts the field
func (field *Field) StartField() error {
	if !field.AllTeamsOnField() {
		return errors.New("the match is not ready, not all driverStations have connected")
	}
	if field.MatchState == STARTED {
		return errors.New("the match has already started, setup the match before you restart it")
	}
	field.TimeLeft = TransitionLength + TeleopLength + EndgameLength + AutoLength
	field.MatchState = STARTED
	return nil
}

// Stops the field
func (field *Field) StopField(isEarly bool) error {
	if field.MatchState != STARTED {
		return errors.New("no matches have started")
	}
	if isEarly {
		field.MatchState = DONE
	} else {
		// TODO: Switch this to initiate a review
		field.MatchState = DONE
	}
	return nil
}

// Get a driverstation by it's team number
func (field *Field) GetDriverStationByTeamNum(teamNum int) *DriverStation {
	if val, ok := field.TeamNumberToDriverStation[teamNum]; ok {
		return val
	}
	return nil
}

// Get a driverstation by it's IP
func (field *Field) GetDriverStationByIP(ip net.Addr) *DriverStation {
	for _, driverStation := range field.TeamNumberToDriverStation {
		if driverStation.TCPSocket.RemoteAddr() == ip {
			return driverStation
		}
	}
	return nil
}

// Get a driverstation by it's team number
func (field *Field) GetAllianceStationFromTeamNum(teamNum int) AllianceStation {
	for allianceStation, team := range field.AllianceStationToTeam {
		if team == teamNum {
			return allianceStation
		}
	}
	return 0 // Returns RED1
}

// Check if a team is in the match
func (field *Field) IsTeamInMatch(teamNum int) bool {
	for _, team := range field.AllianceStationToTeam {
		if team == teamNum {
			return true
		}
	}
	return false
}

// Disable all robots
func (field *Field) DisableAllRobots() {
	field.MatchState = PAUSED
	for _, driverStation := range field.TeamNumberToDriverStation {
		driverStation.Enabled = false
	}
}

// Enable all robots
func (field *Field) EnableAllRobots() {
	field.MatchState = STARTED
	for _, teamNum := range field.AllianceStationToTeam {
		driverStation, ok := field.TeamNumberToDriverStation[teamNum]
		if !ok {
			continue
		}
		driverStation.Enabled = true
	}
}

// Kicks all driverstations from the FMS
func (field *Field) KickAllDriverStations() {
	for _, driverStation := range field.TeamNumberToDriverStation {
		driverStation.Kick()
	}
}

// Checks if all teams are online
func (field *Field) AllTeamsOnField() bool {
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
	return hasAllTeamsOnField
}

// This is the timer for the field, it ticks every second
func (field *Field) fieldTimer() {
	for {
		if field.MatchState == STARTED {
			if field.TimeLeft > TransitionLength + TeleopLength + EndgameLength {
				if field.CurrentPhase != AUTONOMOUS {
					go PlayWAV("audio/CHARGE.wav", time.Millisecond * 1500)
				}
				field.CurrentPhase = AUTONOMOUS
			} else if field.TimeLeft > TeleopLength + EndgameLength {
				if field.CurrentPhase != TRANSITION {
					go PlayWAV("audio/ENDMATCH.wav", time.Millisecond * 500)
				}
				field.CurrentPhase = TRANSITION
			} else if field.TimeLeft > EndgameLength {
				if field.CurrentPhase != TELEOP {
					go PlayWAV("audio/three-bells.wav", time.Millisecond * 1500)
				}
				field.CurrentPhase = TELEOP
			} else if field.TimeLeft > 0 {
				if field.CurrentPhase != ENDGAME {
					go PlayWAV("audio/warning.wav", time.Millisecond * 3000)
				}
				field.CurrentPhase = ENDGAME
			} else {
				go PlayWAV("audio/ENDMATCH.wav", time.Millisecond * 1000)
				err := field.StopField(false)
				if err != nil {
					continue
				}
				continue
			}
			field.TimeLeft--
		}
		time.Sleep(time.Second)
	}
}

// This is the field's tick loop, it ticks every 500 ms
func (field *Field) tick() {
	for {
		// Check if all teams are on field in order to say that the game is ready.
		if field.AllTeamsOnField() && field.MatchState != STARTED && field.MatchState != DONE &&  field.MatchState != PAUSED && field.MatchState != INREVIEW {
			field.MatchState = READY
		}
		for _, driverStation := range field.TeamNumberToDriverStation {
			driverStation.tick()
		}
		time.Sleep(time.Millisecond * 500)
	}
}

// Listens for TCP connections from Driverstations
func (field *Field) listenTCP() {
	listener, err := net.Listen("tcp", "10.0.100.5:1750")
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

		go field.handleTCPConnection(conn)
	}
}

// Handles every TCP connection to the FMS
func (field *Field) handleTCPConnection(conn net.Conn) {
	// errorNum is used to detect if the tcpConn had closed remotely. TODO: Check if there is a better way to do this.
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
}

// Listens for UDP messages from Driverstations
func (field *Field) listenUDP() {
	udpAddr, err := net.ResolveUDPAddr("udp4", "10.0.100.5:1160")
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
		field.handleUDPMessage(bytes)
	}
}

func (field *Field) handleUDPMessage(bytes [50]byte) {
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