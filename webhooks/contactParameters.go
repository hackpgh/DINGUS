package webhooks

import (
	"encoding/json"
	"fmt"
)

type ContactParameters struct {
	ContactId        string `json:"Contact.Id"`
	ProfileChanged   string `json:"ProfileChanged"`
	Action           string `json:"Action"`
	ProfileChangedBy string `json:"ProfileChangedBy"`
}

func (cp *ContactParameters) Validate() error {
	if len(cp.ContactId) <= 0 {
		return fmt.Errorf("ContactId is required")
	}
	return nil
}

func (cp *ContactParameters) String() string {
	return fmt.Sprintf("ContactId: %s",
		cp.ContactId)
}

func (cp *ContactParameters) ToJSON() ([]byte, error) {
	return json.Marshal(cp)
}
