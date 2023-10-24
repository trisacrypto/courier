package api

import "context"

type CourierClient interface {
	Status(context.Context) (*StatusReply, error)
}

// Reply encodes generic JSON responses from the API.
type Reply struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type StatusReply struct {
	Status  string `json:"status"`
	Uptime  string `json:"uptime,omitempty"`
	Version string `json:"version,omitempty"`
}
