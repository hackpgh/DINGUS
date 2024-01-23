package webhooks

import (
	"encoding/json"
	"errors"
)

// NewParameters creates an instance of Parameter based on the webhook type.
func NewParameters(messageType string, rawData json.RawMessage) (Parameter, error) {
	switch messageType {
	case "ContactModified":
		var cp ContactParameters
		if err := json.Unmarshal(rawData, &cp); err != nil {
			return nil, err
		}
		return &cp, nil
	case "Membership":
		var mp MembershipParameters
		if err := json.Unmarshal(rawData, &mp); err != nil {
			return nil, err
		}
		return &mp, nil
	default:
		return nil, errors.New("unsupported message type")
	}
}
