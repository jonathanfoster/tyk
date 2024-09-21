package kv

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/smithy-go"
)

// DummySecretsManagerClient is an in-memory implementation of the SecretsManagerClient interface.
type DummySecretsManagerClient struct {
	secrets map[string]string
}

// NewDummySecretsManagerClient creates a new DummySecretsManagerClient with the given secrets.
func NewDummySecretsManagerClient(secrets map[string]string) *DummySecretsManagerClient {
	return &DummySecretsManagerClient{
		secrets: secrets,
	}
}

// GetSecretValue gets the value of a secret.
func (c *DummySecretsManagerClient) GetSecretValue(
	_ context.Context,
	params *secretsmanager.GetSecretValueInput,
	_ ...func(*secretsmanager.Options),
) (*secretsmanager.GetSecretValueOutput, error) {
	value, ok := c.secrets[*params.SecretId]
	if !ok {
		return nil, &smithy.GenericAPIError{
			Code:    "ResourceNotFoundException",
			Message: "The requested secret was not found.",
		}
	}
	output := &secretsmanager.GetSecretValueOutput{
		Name:         params.SecretId,
		SecretString: aws.String(value),
	}
	return output, nil
}
