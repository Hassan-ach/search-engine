package model

type Word struct {
	Word string  `json:"word"`
	Idf  float64 `json:"idf"`
	Tf   int     `json:"tf"`
}
