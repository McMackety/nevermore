package field

// The length of the autonomous period in seconds
var AutoLength = 15

// The length of the transition period in seconds
var TransitionLength = 1

// The length of the teleop period in seconds
var TeleopLength = 105

// The length of the endgame period in seconds
var EndgameLength = 30

// Returns the time formatted for use in a GUI or the driverstation
func GetFormattedTime(time int) int {
	if time > TransitionLength+TeleopLength+EndgameLength {
		return time - (TransitionLength + TeleopLength + EndgameLength)
	} else if time > TeleopLength+EndgameLength {
		return 0
	} else if time > 0 {
		return time
	} else {
		return 0
	}
}
