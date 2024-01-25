package webhooks

import (
	"encoding/json"
	"fmt"
)

// RenewalStrategy is a custom type to represent the renewal strategy of the membership level.
type RenewalStrategy int

// Enum-like constants for RenewalStrategy.
const (
	StrategyNever      RenewalStrategy = 1
	StrategyMonthly    RenewalStrategy = 3
	StrategyQuarterly  RenewalStrategy = 4
	StrategyTwiceAYear RenewalStrategy = 5
	StrategyYearly     RenewalStrategy = 6
)

// LevelType represents the type of membership level.
type LevelType int

const (
	TypeBundle     LevelType = 1
	TypeIndividual LevelType = 2
)

type MembershipLevelParameters struct {
	Action          string          `json:"Action"`
	LevelId         int             `json:"Level.Id"`
	MembershipFee   float64         `json:"Level.MembershipFee"`
	RenewalStrategy RenewalStrategy `json:"Level.RenewalStrategy"`
	Title           string          `json:"Level.Title"`
	Type            LevelType       `json:"Level.Type"`
}

func (mlp *MembershipLevelParameters) Validate() error {
	switch mlp.Action {
	case "Created", "Disabled", "PriceChanged", "RenewalStrategyChanged", "TitleChanged":
		// valid action
	default:
		return fmt.Errorf("invalid Action: %s", mlp.Action)
	}

	return nil
}

func (mlp *MembershipLevelParameters) String() string {
	return fmt.Sprintf("Action: %s, LevelId: %d, MembershipFee: %.2f, RenewalStrategy: %d, Title: %s, Type: %d",
		mlp.Action, mlp.LevelId, mlp.MembershipFee, mlp.RenewalStrategy, mlp.Title, mlp.Type)
}

func (mlp *MembershipLevelParameters) ToJSON() ([]byte, error) {
	return json.Marshal(mlp)
}
