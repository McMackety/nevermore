package field

var AutoLength = 15

var TransitionLength = 1

var TeleopLength = 105

var EndgameLength = 30

func GetFormattedTime(time int) int {
	if time > TransitionLength + TeleopLength + EndgameLength {
		return time - TransitionLength + TeleopLength + EndgameLength
	} else if time > TeleopLength + EndgameLength {
		return time
	} else if time > EndgameLength {
		return time
	} else if time > 0 {
		return time
	} else {
		return 0
	}
}