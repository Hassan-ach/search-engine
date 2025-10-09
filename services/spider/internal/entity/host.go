package entity

import "spider/internal/utils"

type Host struct {
	MaxRetry       int
	MaxPages       int
	Delay          int
	Name           string
	AllowedUrls    []string
	NotAllwedPaths []string
	DiscovedURLs   *utils.SetQueu[string]
	VisitedURLs    *utils.Set[string]
}
