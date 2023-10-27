package secrets

// SecretsOption allows us to configure the secrets client when it is created.
type SecretsOption func(s *GoogleSecrets) error

func WithGRPCClient(client GRPCSecretClient) SecretsOption {
	return func(s *GoogleSecrets) error {
		s.client = client
		return nil
	}
}
