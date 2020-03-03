package field

import (
	"fmt"
	"github.com/McMackety/nevermore/web"
	"net"
	"time"
)

type DriverStation struct {
	TCPSocket net.Conn `json:"-"`
	CurrentField *Field `json:"-"`
	TeamNumber int `json:"teamNum"`
	EmergencyStopped bool `json:"eStop"`
	Comms bool `json:"comms"`
	RadioPing bool `json:"radioPing"`
	RioPing bool `json:"rioPing"`
	Enabled bool `json:"enabled"`
	Mode Mode `json:"mode"`
	BatteryVoltage float64 `json:"batteryVoltage"`
	UDPSequenceNum int `json:"-"`
	LostPacketsNum int `json:"-"`
	AverageTripTime int `json:"-"`
	Station AllianceStation `json:"allianceStation"`
	Status Status `json:"status"`
	LastUDPMessage time.Time `json:"-"`
	UDPConn net.Conn `json:"-"`
}

func (driverStation *DriverStation) sendEventName() {
	data := []byte{
		0x14,
		byte(len(driverStation.CurrentField.EventName)),
	}

	data = append(data, []byte(driverStation.CurrentField.EventName)...)
	driverStation.TCPSocket.Write(prefixWithSize(data))
}

func (driverStation *DriverStation) sendStationInfo() {
	data := []byte{
		0x19,
		byte(driverStation.Station),
		byte(driverStation.Status),
	}

	driverStation.TCPSocket.Write(prefixWithSize(data))
}



func (driverStation *DriverStation) tick() {
	if time.Since(driverStation.LastUDPMessage).Seconds() > 2 {
		driverStation.Kick()
	} else {
		// Update all Web Clients for updates every tick.
		// Uses "driverStationTick_{teamNum} as Event Name
		web.WebEventsServer.EmitJSONAll("driverStationTick_" + string(driverStation.TeamNumber), driverStation)

		var packet [22]byte
		packet[0] = byte(driverStation.UDPSequenceNum >> 8 & 0xff)
		packet[1] = byte(driverStation.UDPSequenceNum & 0xff)

		packet[2] = 0

		packet[3] = 0
		if driverStation.Mode == AUTONOMOUSMODE {
			packet[3] |= 0x02
		}
		if driverStation.Enabled {
			packet[3] |= 0x04
		}
		if driverStation.EmergencyStopped {
			packet[3] |= 0x80
		}

		packet[4] = 0 // Unknown

		packet[5] = byte(driverStation.Station)

		packet[6] = byte(CurrentField.MatchLevel)

		packet[7] = byte(CurrentField.MatchNumber >> 8 & 0xff)

		packet[8] = byte(CurrentField.MatchNumber & 0xff)

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

func (driverStation *DriverStation) Kick() {
	println("Kicked")
	web.WebEventsServer.EmitJSONAll("driverStationKicked", driverStation.TeamNumber)
	driverStation.TCPSocket.Close()
	driverStation.UDPConn.Close()
}

func (driverStation *DriverStation) receiveUDP(eStop bool, comms bool, radioPing bool, rioPing bool, enabled bool, mode Mode, batteryVoltage float64) {
	println("received")
	driverStation.LastUDPMessage = time.Now()
	fmt.Println(driverStation.LastUDPMessage)
	driverStation.EmergencyStopped = eStop
	driverStation.Comms = comms
	driverStation.RadioPing = radioPing
	driverStation.RioPing = rioPing
	driverStation.Enabled = enabled
	driverStation.Mode = mode
	driverStation.BatteryVoltage = batteryVoltage
}

func prefixWithSize(bytes []byte) []byte {
	tempBuf := []byte{
		byte(len(bytes) >> 8 & 0xff),
		byte(len(bytes) & 0xff),
	}
	return append(tempBuf, bytes...)
}