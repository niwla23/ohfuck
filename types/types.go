package types

import "time"

type MonitorState struct {
	Name           string    `json:"name"`
	FriendlyName   string    `json:"friendlyName"`
	Up             bool      `json:"up"`
	Reason         string    `json:"reason"`
	LastReportTime time.Time `json:"lastReportTime"`
}
