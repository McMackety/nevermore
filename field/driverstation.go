package field

import (
	"net"
	"time"
)

type DriverStation struct {
	TCPSocket        net.Conn        `json:"-"`
	CurrentField     *Field          `json:"-"`
	TeamNumber       int             `json:"teamNum"`
	EmergencyStopped bool            `json:"eStop"`
	RequestEmergencyStop bool        `json:"requestEStop"`
	Comms            bool            `json:"comms"`
	RadioPing        bool            `json:"radioPing"`
	RioPing          bool            `json:"rioPing"`
	RequestEnabled   bool            `json:"requestEnabled"`
	Enabled          bool            `json:"enabled"`
	BatteryVoltage   float64         `json:"batteryVoltage"`
	UDPSequenceNum   int             `json:"-"`
	Station          AllianceStation `json:"allianceStation"`
	Status           Status          `json:"status"`
	LastUDPMessage   time.Time       `json:"-"`
	UDPConn          net.Conn        `json:"-"`
}

// Creates a new driver station connection.
func (field *Field) createDriverStation(teamNum int, socket net.Conn, udpSocket net.Conn) {
	driverStation := &DriverStation{
		TCPSocket:        socket,
		CurrentField:     field,
		TeamNumber:       teamNum,
		EmergencyStopped: false,
		RequestEmergencyStop: false,
		Comms:            false,
		RadioPing:        false,
		RioPing:          false,
		Enabled:          true,
		BatteryVoltage:   0.0,
		UDPSequenceNum:   0,
		LastUDPMessage:   time.Now(),
		UDPConn:          udpSocket,
	}

	field.TeamNumberToDriverStation[teamNum] = driverStation

	if !field.IsTeamInMatch(teamNum) {
		driverStation.Status = WAITING
	} else {
		driverStation.Status = GOOD
		driverStation.Station = field.GetAllianceStationFromTeamNum(teamNum)
	}

	// Send Event and Station Info
	driverStation.SendStationInfo()
	driverStation.SendEventName()
}

// Returns true if in autonomous period, false if not.
func (driverStation *DriverStation) IsInAutonomous() bool {
	if driverStation.CurrentField.TimeLeft > TransitionLength+TeleopLength+EndgameLength {
		return true
	}
	return false
}

// Returns true if in autonomous period, false if not.
func (driverStation *DriverStation) ShouldBeEnabled() bool {
	if driverStation.CurrentField.MatchState != STARTED {
		return false
	}
	if driverStation.CurrentField.TimeLeft > TransitionLength+TeleopLength+EndgameLength {
		return true
	} else if driverStation.CurrentField.TimeLeft > TeleopLength+EndgameLength {
		return false
	} else if driverStation.CurrentField.TimeLeft > 0 {
		return true
	}
	return false
}

// Sends the name of the event
func (driverStation *DriverStation) SendEventName() {
	data := []byte{
		0x14,
		byte(len(driverStation.CurrentField.EventName)),
	}

	data = append(data, []byte(driverStation.CurrentField.EventName)...)
	driverStation.TCPSocket.Write(prefixWithSize(data))
}

// Sends the station's info
func (driverStation *DriverStation) SendStationInfo() {
	data := []byte{
		0x19,
		byte(driverStation.Station),
		byte(driverStation.Status),
	}

	driverStation.TCPSocket.Write(prefixWithSize(data))
}

// Kicks the driverstation
func (driverStation *DriverStation) Kick() {
	delete(driverStation.CurrentField.TeamNumberToDriverStation, driverStation.TeamNumber)
	driverStation.TCPSocket.Close()
	driverStation.UDPConn.Close()
}

// Ticks the driverstation, ran every 500 ms
func (driverStation *DriverStation) tick() {
	if time.Since(driverStation.LastUDPMessage).Seconds() > 2 {
		driverStation.Kick()
	} else {
		// Update all Web Clients for updates every tick.
		// Uses "driverStationTick_{teamNum} as Event Name
		enabled := driverStation.Enabled
		if !driverStation.ShouldBeEnabled() {
			if driverStation.CurrentField.MatchLevel != MATCHTEST {
				enabled = false
			}
		}

		autonomous := driverStation.IsInAutonomous()

		var packet [22]byte
		packet[0] = byte(driverStation.UDPSequenceNum >> 8 & 0xff)
		packet[1] = byte(driverStation.UDPSequenceNum & 0xff)

		packet[2] = 0

		packet[3] = 0
		if autonomous {
			packet[3] |= 0x02
		}
		if enabled {
			packet[3] |= 0x04
		}
		if driverStation.EmergencyStopped {
			packet[3] |= 0x80
		}

		packet[4] = 0 // Unknown

		packet[5] = byte(driverStation.Station)

		packet[6] = byte(driverStation.CurrentField.MatchLevel)

		packet[7] = byte(driverStation.CurrentField.MatchNumber >> 8 & 0xff)

		packet[8] = byte(driverStation.CurrentField.MatchNumber & 0xff)

		packet[9] = 1 // Useless Replay Number (To Us)

		// Current time.
		currentTime := time.Now()
		packet[10] = byte(((currentTime.Nanosecond() / 1000) >> 24) & 0xff)
		packet[11] = byte(((currentTime.Nanosecond() / 1000) >> 16) & 0xff)
		packet[12] = byte(((currentTime.Nanosecond() / 1000) >> 8) & 0xff)
		packet[13] = byte((currentTime.Nanosecond() / 1000) & 0xff)
		packet[14] = byte(currentTime.Second())
		packet[15] = byte(currentTime.Minute())
		packet[16] = byte(currentTime.Hour())
		packet[17] = byte(currentTime.Day())
		packet[18] = byte(currentTime.Month())
		packet[19] = byte(currentTime.Year() - 1900)

		packet[20] = byte(GetFormattedTime(driverStation.CurrentField.TimeLeft) >> 8 & 0xff)
		packet[21] = byte(GetFormattedTime(driverStation.CurrentField.TimeLeft) & 0xff)

		driverStation.UDPConn.Write(packet[:])

		driverStation.UDPSequenceNum++
	}
}

// Called whenever a UDP message was received
func (driverStation *DriverStation) receiveUDP(eStop bool, comms bool, radioPing bool, rioPing bool, enabled bool, mode Mode, batteryVoltage float64) {
	driverStation.LastUDPMessage = time.Now()
	driverStation.Comms = comms
	driverStation.RadioPing = radioPing
	driverStation.RioPing = rioPing
	driverStation.BatteryVoltage = batteryVoltage
	driverStation.RequestEnabled = enabled
	driverStation.RequestEmergencyStop = eStop
}

func prefixWithSize(bytes []byte) []byte {
	tempBuf := []byte{
		byte(len(bytes) >> 8 & 0xff),
		byte(len(bytes) & 0xff),
	}
	return append(tempBuf, bytes...)
}
