package field

import (
	"fmt"
	"log"
	"net"
	"time"
)

var correctNetwork bool = false
var CurrentField Field

type Field struct {
	TeamNumberToDriverStation map[int]DriverStation
	AllianceStationToTeam map[AllianceStation]int
	MatchPaused bool
	MatchStarted bool
	MatchReady bool
	MatchStartedAt time.Time
	MatchLevel Level
	MatchNumber int
	TimeLeft int
	UDPSocket *net.UDPConn
	EventName string
	CurrentPhase string
}

// InitField starts the networking for the FMS
func InitField() {
	// Check for a network interface with 10.0.100.5
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if addr.String() == "10.0.100.5" {
		 	correctNetwork = true
		}
	}
	if !correctNetwork {
		log.Fatalln("Couldn't find network interface with an IP of 10.0.100.5, please read the setup guide at https://github.com/McMackety/nevermore/wiki!")
	}

	CurrentField = Field{
		MatchPaused: false,
		MatchStarted: false,
		MatchReady: false,
		MatchStartedAt: nil,
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
	for _, driverStation := range field.TeamNumberToDriverStation {
		driverStation.Enabled = true
	}
}

func (field *Field) KickAllDriverStations() {
	for teamNum, driverStation := range field.TeamNumberToDriverStation {
		delete(field.TeamNumberToDriverStation, teamNum)
		driverStation.Kick()
	}
}

func (field *Field) SetupField(matchNum int, tournamentLevel Level, red1 int, red2 int, red3 int, blue1 int, blue2 int, blue3 int) {
	field.KickAllDriverStations()
	field.MatchStarted = false
	field.MatchLevel = tournamentLevel
	field.AllianceStationToTeam[REDSTATION1] = red1
	field.AllianceStationToTeam[REDSTATION2] = red2
	field.AllianceStationToTeam[REDSTATION3] = red3
	field.AllianceStationToTeam[BLUESTATION1] = blue1
	field.AllianceStationToTeam[BLUESTATION2] = blue2
	field.AllianceStationToTeam[BLUESTATION3] = blue3
}

func (field *Field) fieldTimer() {
	if field.MatchStarted && !field.MatchPaused {
		field.TimeLeft--
		// INFINITE RECHARGE START
		if field.TimeLeft > TransitionLength + TeleopLength + EndgameLength {
			if field.CurrentPhase != "auto" {
				field.SetAllRobotModes(AUTONOMOUSMODE)
				field.EnableAllRobots()
			}
			field.CurrentPhase = "auto"
		} else if field.TimeLeft > TeleopLength + EndgameLength {
			if field.CurrentPhase != "transition" {
				field.DisableAllRobotsNoPause()
			}
			field.CurrentPhase = "transition"
		} else if field.TimeLeft > EndgameLength {
			if field.CurrentPhase != "teleop" {
				field.SetAllRobotModes(TELEOPMODE)
				field.EnableAllRobots()
			}
			field.CurrentPhase = "teleop"
		} else if field.TimeLeft > 0 {
			field.CurrentPhase = "endgame"
		} else {
			field.DisableAllRobotsNoPause()
		}

		// INFINITE RECHARGE END
	}
	time.Sleep(time.Second)
	field.fieldTimer()
}

func (field *Field) tick() {
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
		Station:          field.GetAllianceStationFromTeamNum(teamNum),
		UDPConn: 		  udpSocket,
	}

	if _, ok := field.TeamNumberToDriverStation[teamNum]; ok {
		driverStation.Kick()
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
	listener, err := net.Listen("tcp", "10.0.100.5:1750")
	if err != nil {
		log.Panicln("Couldn't start FMS TCP Server: " + err.Error())
	}
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
				var bytes []byte
				_, err := conn.Read(bytes)
				if err != nil {
					errorNum++
					continue
				}
				if bytes[2] == 0x18 {
					teamNum := int((bytes[3] << 8) + bytes[4])
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
	udpAddr, err := net.ResolveUDPAddr("udp", "10.0.100.5:1160")
	if err != nil {
		log.Panicln("Bad Address for UDP: " + err.Error())
	}

	listener, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		log.Panicln("Couldn't start FMS UDP Server: " + err.Error())
	}

	field.UDPSocket = listener

	defer listener.Close()

	for {
		var bytes []byte
		listener.Read(bytes)
		eStopped := (bytes[3] >> 7 & 0x01) == 1
		comms := (bytes[3] >> 5 & 0x01) == 1
		radioPing := (bytes[3] >> 4 & 0x01) == 1
		rioPing := (bytes[3] >> 3 & 0x01) == 1
		enabled := (bytes[3] >> 2 & 0x01) == 1
		mode := Mode(bytes[3] & 0x03)
		teamNum := int((bytes[4] << 8) + bytes[5])
		batteryVoltage := float64(bytes[6]) + float64(bytes[7])/256

		if driverStation := field.GetDriverStationByTeamNum(teamNum); driverStation != nil {
			driverStation.receiveUDP(eStopped, comms, radioPing, rioPing, enabled, mode, batteryVoltage)
		}
	}
}