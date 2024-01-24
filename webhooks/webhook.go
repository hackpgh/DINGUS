package webhooks

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Webhook struct {
	AccountId   int    `json:"AccountId"`
	MessageType string `json:"MessageType"`
	Parameters  Parameter
}

// UnmarshalJSON is a custom unmarshaler for Webhook that handles different Parameter types.
func (w *Webhook) UnmarshalJSON(data []byte) error {
	var raw struct {
		AccountId   string          `json:"AccountId"`
		MessageType string          `json:"MessageType"`
		Parameters  json.RawMessage `json:"Parameters"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	w.AccountId, _ = strconv.Atoi(raw.AccountId)
	w.MessageType = raw.MessageType

	param, err := NewParameters(raw.MessageType, raw.Parameters)
	if err != nil {
		return fmt.Errorf("failed to create parameters: %v", err)
	}

	w.Parameters = param
	return nil
}

func (w *Webhook) String() string {
	return fmt.Sprintf("AccountId: %d, MessageType: %s, Parameters: %v", w.AccountId, w.MessageType, w.Parameters)
}
