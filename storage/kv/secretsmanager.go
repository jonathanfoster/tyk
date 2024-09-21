package kv

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/smithy-go"
	"github.com/pkg/errors"

	"github.com/TykTechnologies/tyk/config"
)

// SecretsManagerClient is an interface for the AWS Secrets Manager client.
type SecretsManagerClient interface {
	// GetSecretValue retrieves a secret from AWS Secrets Manager.
	GetSecretValue(
		ctx context.Context,
		params *secretsmanager.GetSecretValueInput,
		optFns ...func(*secretsmanager.Options),
	) (*secretsmanager.GetSecretValueOutput, error)
}

// SecretsManager is an implementation of a KV store which uses AWS Secrets Manager as its backend.
type SecretsManager struct {
	client SecretsManagerClient
}

// NewSecretsManager returns an AWS Secrets Manager KV store with default configuration.
func NewSecretsManager() (*SecretsManager, error) {
	return NewSecretsManagerWithConfig(config.SecretsManagerConfig{})
}

// NewSecretsManagerWithConfig returns an AWS Secrets Manager KV store with a custom configuration.
func NewSecretsManagerWithConfig(conf config.SecretsManagerConfig) (*SecretsManager, error) {
	cfg, err := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsconfig.WithRegion(conf.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(conf.AccessKeyID, conf.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error loading default config")
	}

	client := secretsmanager.NewFromConfig(cfg)

	return &SecretsManager{client: client}, nil
}

// NewSecretsManagerWithClient returns a configured AWS Secrets Manager KV store with a custom client.
func NewSecretsManagerWithClient(client SecretsManagerClient) *SecretsManager {
	return &SecretsManager{client: client}
}

// Get retrieves a secret from AWS Secrets Manager. Key is the secret name if the secret is a plain text value (e.g.,
// "config") or the secret name followed by the key name if the secret is a JSON document (e.g., "config.node_secret").
func (s *SecretsManager) Get(key string) (string, error) {
	secretId, secretKey := splitKVPathAndKey(key)

	value, err := s.client.GetSecretValue(
		context.TODO(),
		&secretsmanager.GetSecretValueInput{SecretId: aws.String(secretId)},
	)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "ResourceNotFoundException" {
				return "", ErrKeyNotFound
			}
		}

		return "", fmt.Errorf("error getting secret value: %w", err)
	}

	if secretKey == "" {
		return *value.SecretString, nil
	}

	jsonValue := make(map[string]interface{})
	err = json.Unmarshal([]byte(*value.SecretString), &jsonValue)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling secret string: %w", err)
	}

	val, ok := jsonValue[secretKey]
	if !ok {
		return "", ErrKeyNotFound
	}

	return fmt.Sprintf("%v", val), nil
}

func splitKVPathAndKey(raw string) (path string, key string) {
	sep := "."
	parts := strings.Split(raw, sep)
	if len(parts) > 1 {
		return parts[0], strings.Join(parts[1:], sep)
	}

	return raw, ""
}
