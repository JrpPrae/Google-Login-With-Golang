package models

type GoogleInformationResponse struct {
	PID        string `json:"pid"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	BirthDate  string `json:"birthdate"`
}
