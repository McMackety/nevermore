package field

// Alliance is the Red/Blue alliance.
type Alliance int

// The blue and red alliances
const (
	RED Alliance = iota
	BLUE
)

// AllianceStation defines the alliance stations (should be obvious)
type AllianceStation int

// The different alliance stations
const (
	RED1 AllianceStation = iota
	RED2
	RED3
	BLUE1
	BLUE2
	BLUE3
)

// Status is the status of the robot
type Status int

// The different statuses for the robots
const (
	GOOD Status = iota
	BAD
	WAITING
)

// Mode is the mode of the robot
type Mode int

// The different modes for the robot
const (
	TELEOPMODE Mode = iota
	TESTMODE
	AUTONOMOUSMODE
)

// Level is the tournament level.
type Level int

// The different tournament levels.
const (
	MATCHTEST Level = iota
	PRACTICE
	QUALIFICATION
	PLAYOFF
)

// State is the currentState of the field.
type State int

// The different field states
const (
	NOTREADY State = iota
	READY
	STARTED
	PAUSED
	INREVIEW
	DONE
)

// Phase is the different phases the field can be in.
type Phase int

// The different field phases.
const (
	NOTHING Phase = iota
	AUTONOMOUS
	TRANSITION
	TELEOP
	ENDGAME
)