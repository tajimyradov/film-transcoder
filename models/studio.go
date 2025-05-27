package models

type Studio struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Abbreviated string `json:"abbreviated"`
}
