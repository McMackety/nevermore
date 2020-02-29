package field

type Alliance int

const (
	REDALLIANCE Alliance = iota
	BLUEALLIANCE
)

type AllianceStation int

const (
	REDSTATION1 AllianceStation = iota
	REDSTATION2
	REDSTATION3
	BLUESTATION1
	BLUESTATION2
	BLUESTATION3
)

type Status int

const (
	GOODSTATUS Status = iota
	BADSTATUS
	WAITINGSTATUS
)

type Mode int

const (
	TELEOPMODE Mode = iota
	TESTMODE
	AUTONOMOUSMODE
)

type Level int

const (
	MATCHTESTLEVEL Level = iota
	PRACTICELEVEL
	QUALIFICATIONLEVEL
	PLAYOFFLEVEL
)