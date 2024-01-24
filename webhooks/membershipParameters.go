package webhooks

import (
	"encoding/json"
	"fmt"
)

// MembershipStatus is a custom type to represent the status of membership.
type MembershipStatus string

// Enum-like constants for MembershipStatus.
const (
	StatusNOOP           MembershipStatus = "0"
	StatusActive         MembershipStatus = "1"
	StatusLapsed         MembershipStatus = "2"
	StatusPendingRenewal MembershipStatus = "3"
	StatusPendingNew     MembershipStatus = "20"
	StatusPendingUpgrade MembershipStatus = "30"
)

// MembershipParameters defines the structure for membership webhook parameters.
type MembershipParameters struct {
	Action            string           `json:"Action"`
	ContactId         string           `json:"Contact.Id"`
	MembershipLevelId string           `json:"Membership.LevelId"`
	MembershipStatus  MembershipStatus `json:"Membership.Status"`
}

func (mp *MembershipParameters) Validate() error {
	var err error
	switch mp.Action {
	case "Enabled", "Disabled", "StatusChanged", "RenewalDateChanged", "LevelChanged":
		break
	default:
		return fmt.Errorf("invalid Action: %s %v", mp.Action, err)
	}

	switch mp.MembershipStatus {
	case StatusNOOP, StatusActive, StatusLapsed, StatusPendingNew, StatusPendingRenewal, StatusPendingUpgrade:
		break
	default:
		return fmt.Errorf("invalid MembershipStatus: %s %v", mp.MembershipStatus, err)
	}

	return nil
}

func (mp *MembershipParameters) String() string {
	return fmt.Sprintf("Action: %s, ContactId: %s, MembershipLevelId: %s, MembershipStatus: %s",
		mp.Action, mp.ContactId, mp.MembershipLevelId, mp.MembershipStatus)
}

func (mp *MembershipParameters) ToJSON() ([]byte, error) {
	return json.Marshal(mp)
}
