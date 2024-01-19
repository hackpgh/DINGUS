package models

type WildApricotWebhook struct {
	AccountId   string `json:"AccountId"`
	MessageType string `json:"MessageType"`
	Parameters  struct {
		ContactId        string `json:"Contact.Id"`
		ProfileChanged   string `json:"ProfileChanged"`
		Action           string `json:"Action"`
		ProfileChangedBy string `json:"ProfileChangedBy"`
	} `json:"Parameters"`
}
