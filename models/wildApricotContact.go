package models

// Contact represents the structure of a contact in the Wild Apricot API's /Contacts response.
type Contact struct {
	FirstName              string       `json:"FirstName"`
	LastName               string       `json:"LastName"`
	Email                  string       `json:"Email"`
	DisplayName            string       `json:"DisplayName"`
	Organization           string       `json:"Organization"`
	ProfileLastUpdated     string       `json:"ProfileLastUpdated"`
	FieldValues            []FieldValue `json:"FieldValues"`
	Id                     int          `json:"Id"`
	Url                    string       `json:"Url"`
	IsAccountAdministrator bool         `json:"IsAccountAdministrator"`
	TermsOfUseAccepted     bool         `json:"TermsOfUseAccepted"`
}

// FieldValue represents the structure for field values in a contact.
type FieldValue struct {
	FieldName  string      `json:"FieldName"`
	Value      interface{} `json:"Value"`
	SystemCode string      `json:"SystemCode"`
}

// Training represents a training item in FieldValue containing active values from a WA multiple choice list.
type SafetyTraining struct {
	Id    int    `json:"Id"`
	Label string `json:"Label"`
}
