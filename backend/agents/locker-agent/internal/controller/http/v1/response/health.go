package response

import "time"

type Health struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

type ServiceInfo struct {
	Service string `json:"service"`
	Version string `json:"version"`
	Status  string `json:"status"`
}
