package secrets

import (
	"context"
	"errors"
	"fmt"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/trisacrypto/courier/pkg/config"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewClient creates a secret manager client from the configuration.
func NewClient(conf config.SecretsConfig, opts ...SecretsOption) (_ SecretManagerClient, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s := &GoogleSecrets{
		parent: "projects/" + conf.Project,
	}

	// Apply provided options
	for _, opt := range opts {
		if err = opt(s); err != nil {
			return nil, err
		}
	}

	if s.client == nil {
		// Specify credentials path if provided
		opts := []option.ClientOption{}
		if conf.Credentials != "" {
			opts = append(opts, option.WithCredentialsFile(conf.Credentials))
		}

		// Create the client
		if s.client, err = secretmanager.NewClient(ctx, opts...); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// GoogleSecrets implements the secret manager interface.
type GoogleSecrets struct {
	parent string
	client GRPCSecretClient
}

var _ SecretManagerClient = &GoogleSecrets{}

//===========================================================================
// Secret Manager Methods
//===========================================================================

// CreateSecret creates a new secret in the child directory of the parent. Does not
// return an error if the secret already exists.
func (s *GoogleSecrets) CreateSecret(ctx context.Context, name string) (err error) {
	// Build the request.
	req := &secretmanagerpb.CreateSecretRequest{
		Parent:   s.parent,
		SecretId: name,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}

	// Call the API, secret response is discarded to avoid leaking secret data.
	if _, err = s.client.CreateSecret(ctx, req); err != nil {
		// If the API call is malformed, it will hang until the internal context times out
		if errors.Is(err, context.DeadlineExceeded) {
			return err
		}

		// The secret can already exist, which is fine because secrets are versioned
		// and we will always retrieve the latest version.
		serr, ok := status.FromError(err)
		if ok && serr.Code() == codes.AlreadyExists {
			return nil
		}

		// If the error is something else, something went wrong.
		return err
	}
	return nil
}

// AddSecretVersion adds a new secret version to the given secret and the
// provided payload. Returns an error if one occurs.
// Note: to add a secret version, the secret must first be created using CreateSecret.
func (s *GoogleSecrets) AddSecretVersion(ctx context.Context, name string, payload []byte) (err error) {
	secretPath := fmt.Sprintf("%s/secrets/%s", s.parent, name)

	// Build the request.
	req := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secretPath,
		Payload: &secretmanagerpb.SecretPayload{
			Data: payload,
		},
	}

	// Call the API, secret response is discarded to avoid leaking secret data.
	if _, err = s.client.AddSecretVersion(ctx, req); err != nil {
		// If the API call is malformed, it will hang until the internal context times out
		if errors.Is(err, context.DeadlineExceeded) {
			return err
		}

		serr, ok := status.FromError(err)
		if ok {
			switch serr.Code() {
			// If the secret does not exist (e.g. has been deleted or hasn't been created yet)
			// we'll get a Not Found error
			case codes.NotFound:
				return ErrSecretNotFound
			// If the secret exceeds 65KiB we'll get a InvalidArgument error
			case codes.InvalidArgument:
				return ErrPayloadTooLarge
			// If we give the wrong path to the project, we get a Permission Denied error
			case codes.PermissionDenied:
				return ErrPermissionsDenied
			}
		}

		// If the error is something else, something went wrong.
		return err
	}

	return nil
}

// GetLatestVersion returns the payload for the latest version of the given secret,
// if one exists, else an error.
func (s *GoogleSecrets) GetLatestVersion(ctx context.Context, name string) (_ []byte, err error) {
	versionPath := fmt.Sprintf("%s/secrets/%s/versions/latest", s.parent, name)

	// Build the request.
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: versionPath,
	}

	// Call the API.
	result, err := s.client.AccessSecretVersion(ctx, req)
	if err != nil {
		// If the API call is malformed, it will hang until the internal context times out
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}

		serr, ok := status.FromError(err)
		if ok && serr.Code() == codes.NotFound {
			return nil, ErrSecretNotFound
		}

		// If the error is something else, something went wrong.
		return nil, err
	}

	return result.Payload.Data, nil
}

// DeleteSecret deletes the secret with the given the name, and all of its versions.
// Note: this is an irreversible operation. Any service or workload that attempts to
// access a deleted secret receives a Not Found error.
func (s *GoogleSecrets) DeleteSecret(ctx context.Context, secret string) error {
	secretPath := fmt.Sprintf("%s/secrets/%s", s.parent, secret)

	// Build the request.
	req := &secretmanagerpb.DeleteSecretRequest{
		Name: secretPath,
	}

	// Call the API.
	err := s.client.DeleteSecret(ctx, req)
	if err != nil {
		// If the API call is malformed, it will hang until the internal context times out
		if errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		// If the error is something else, something went wrong.
		return err
	}
	return nil
}
