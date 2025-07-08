package models

type Studio struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Abbreviated string `json:"abbreviated"`
}
