package api

// ClientOption allows the API client to be configured when it is created.
type ClientOption func(c *APIv1) error
