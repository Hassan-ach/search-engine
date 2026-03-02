package model

type Edge struct {
	From *Page `json:"from"`
	To   *Page `json:"to"`
}
