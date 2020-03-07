package scoring

import (
	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	"math/rand"
)

type ScoringInterface interface {
	UpdateRedScoringData(redData map[string]interface{})
	UpdateBlueScoringData(blueData map[string]interface{})
	GetScoringDataRed() map[string]interface{}
	GetScoringDataBlue() map[string]interface{}
	GetFinalScore() (redScore int, blueScore int)
	ShouldSendColorRed(currentColor string) bool
	ShouldSendColorBlue(currentColor string) bool
}

func CreateScoringInterface() ScoringInterface {
	return &InfiniteRechargeScoring{
		RedData: InfiniteRechargeScoringData{
			AutoInitiationLine:       0,
			TotalPowerCells:          0,
			AutoLowPowerCells:        0,
			AutoOuterPowerCells:      0,
			AutoInnerPowerCells:      0,
			TeleopLowPowerCells:      0,
			TeleopOuterPowerCells:    0,
			TeleopInnerPowerCells:    0,
			ShieldGeneratorCharge:    0,
			RotationControlCompleted: false,
			PositionControlCompleted: false,
			HangingRobots:            0,
			ParkedRobots:             0,
			LevelSwitch:              false,
			OpponentFoul:             0,
			OpponentTechFoul:         0,
		},
		BlueData: InfiniteRechargeScoringData{
			AutoInitiationLine:       0,
			TotalPowerCells:          0,
			AutoLowPowerCells:        0,
			AutoOuterPowerCells:      0,
			AutoInnerPowerCells:      0,
			TeleopLowPowerCells:      0,
			TeleopOuterPowerCells:    0,
			TeleopInnerPowerCells:    0,
			ShieldGeneratorCharge:    0,
			RotationControlCompleted: false,
			PositionControlCompleted: false,
			HangingRobots:            0,
			ParkedRobots:             0,
			LevelSwitch:              false,
			OpponentFoul:             0,
			OpponentTechFoul:         0,
		},
		RedColor: "",
		BlueColor: "",
	}
}

type InfiniteRechargeScoring struct {
	RedData InfiniteRechargeScoringData
	BlueData InfiniteRechargeScoringData
	RedColor string
	BlueColor string
}

func (scoring *InfiniteRechargeScoring) UpdateRedScoringData(data map[string]interface{}) {
	mapstructure.Decode(data, &scoring.RedData)
}

func (scoring *InfiniteRechargeScoring) UpdateBlueScoringData(data map[string]interface{}) {
	mapstructure.Decode(data, &scoring.BlueData)
}

func (scoring *InfiniteRechargeScoring) GetScoringDataRed() map[string]interface{} {
	return structs.Map(scoring.RedData)
}

func (scoring *InfiniteRechargeScoring) GetScoringDataBlue() map[string]interface{} {
	return structs.Map(scoring.BlueData)
}

func (scoring *InfiniteRechargeScoring) GetFinalScore() (redScore int, blueScore int) {
	return scoring.RedData.calcScore(), scoring.BlueData.calcScore()
}

func (scoring *InfiniteRechargeScoring) ShouldSendColorRed(currentColor string) bool {
	if scoring.RedColor == "" {
		if scoring.RedData.RotationControlCompleted && scoring.RedData.TotalPowerCells >= 9 + 20 {
			scoring.RedColor = getRandomColor(currentColor)
			return true
		}
	}
	return false
}

func (scoring *InfiniteRechargeScoring) ShouldSendColorBlue(currentColor string) bool {
	if scoring.BlueColor == "" {
		if scoring.BlueData.RotationControlCompleted && scoring.BlueData.TotalPowerCells >= 9 + 20 {
			scoring.BlueColor = getRandomColor(currentColor)
			return true
		}
	}
	return false
}

type InfiniteRechargeScoringData struct {
	AutoInitiationLine int
	TotalPowerCells int
	AutoLowPowerCells int
	AutoOuterPowerCells int
	AutoInnerPowerCells int
	TeleopLowPowerCells int
	TeleopOuterPowerCells int
	TeleopInnerPowerCells int
	ShieldGeneratorCharge int
	RotationControlCompleted bool
	PositionControlCompleted bool
	HangingRobots int
	ParkedRobots int
	LevelSwitch bool
	OpponentFoul int
	OpponentTechFoul int
}

func (scoreData *InfiniteRechargeScoringData) calcScore() int {
	score := 0
	score += scoreData.AutoInitiationLine*5
	score += scoreData.AutoInnerPowerCells*6
	score += scoreData.AutoOuterPowerCells*4
	score += scoreData.AutoLowPowerCells*2
	score += scoreData.TeleopInnerPowerCells*3
	score += scoreData.TeleopOuterPowerCells*2
	score += scoreData.TeleopLowPowerCells*1
	if scoreData.RotationControlCompleted {
		score += 10
	}
	if scoreData.PositionControlCompleted {
		score += 20
	}
	score += scoreData.HangingRobots*25
	score += scoreData.ParkedRobots*5
	if scoreData.LevelSwitch && scoreData.HangingRobots > 0 {
		score += 15
	}
	return score
}

func getRandomColor(otherThan string) string {
	for {
		switch rand.Intn(4) {
		case 0:
			if otherThan == "y" {
				continue
			}
			return "y"
		case 1:
			if otherThan == "b" {
				continue
			}
			return "b"
		case 2:
			if otherThan == "g" {
				continue
			}
			return "g"
		case 3:
			if otherThan == "r" {
				continue
			}
			return "r"
		}
	}
}