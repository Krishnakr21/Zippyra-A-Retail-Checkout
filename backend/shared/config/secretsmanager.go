package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type SecretsManagerFetcher struct {
	client      *secretsmanager.Client
	environment string
	serviceName string
}

func NewSecretsManagerFetcher(ctx context.Context, env, service, region string) (*SecretsManagerFetcher, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	client := secretsmanager.NewFromConfig(cfg)
	return &SecretsManagerFetcher{
		client:      client,
		environment: env,
		serviceName: service,
	}, nil
}

func (s *SecretsManagerFetcher) FetchAndApplySecrets(ctx context.Context) (map[string]string, error) {
	secretName := fmt.Sprintf("zippyra/%s/%s/", s.environment, s.serviceName)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := s.client.GetSecretValue(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch secret %s: %w", secretName, err)
	}

	if result.SecretString == nil {
		return nil, fmt.Errorf("secret %s is empty", secretName)
	}

	var secrets map[string]string
	if err := json.Unmarshal([]byte(*result.SecretString), &secrets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret JSON: %w", err)
	}

	for k, v := range secrets {
		_ = os.Setenv(k, v)
	}

	return secrets, nil
}

func StartSecretsManagerHotReload(ctx context.Context, env, service, region string, pollIntervalMinutes int) {
	if env != "staging" && env != "production" {
		return
	}

	fetcher, err := NewSecretsManagerFetcher(ctx, env, service, region)
	if err != nil {
		fmt.Printf("[SECRETS] Failed to initialize SecretsManager fetcher: %v\n", err)
		return
	}

	// Initial fetch
	if _, err := fetcher.FetchAndApplySecrets(ctx); err != nil {
		fmt.Printf("[SECRETS] Initial secrets pull failed: %v\n", err)
	} else {
		fmt.Printf("[SECRETS] Successfully pulled and applied secrets for zippyra/%s/%s/\n", env, service)
	}

	ticker := time.NewTicker(time.Duration(pollIntervalMinutes) * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				fetchCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				if _, err := fetcher.FetchAndApplySecrets(fetchCtx); err != nil {
					fmt.Printf("[SECRETS] Hot-reload secrets poll error: %v\n", err)
				} else {
					fmt.Printf("[SECRETS] Hot-reloaded secrets for zippyra/%s/%s/\n", env, service)
				}
				cancel()
			}
		}
	}()
}
