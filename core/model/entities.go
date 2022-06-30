package model

type Course struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	AccessRestrictedByDate bool   `json:"access_restricted_by_date"`
}
