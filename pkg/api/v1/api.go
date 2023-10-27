package api

import "context"

type CourierClient interface {
	Status(context.Context) (*StatusReply, error)
	StoreCertificatePassword(context.Context, *StorePasswordRequest) error
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

type StorePasswordRequest struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}
