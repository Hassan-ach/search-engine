package model

import (
	"github.com/google/uuid"
)

type Page struct {
	ID          uuid.UUID      `json:"id"`
	URL         string         `json:"url"`
	PRScore     float64        `json:"pr_score"`
	Words       map[string]int `json:"words"`
	GlobalScore float64        `json:"global_score"`
	MetaData    MetaData       `json:"metadata"`
}
