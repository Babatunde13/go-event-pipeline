package config

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	secretsmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type Config struct {
	Brokers                  string `json:"KAFKA_BROKERS"`
	KafkaTopic               string `json:"KAFKA_TOPIC"`
	EventBusName             string `json:"EVENT_BUS_NAME"`
	EventBusSource           string `json:"EVENT_BUS_SOURCE"`
	PrometheusPushGatewayUrl string `json:"PROMETHEUS_PUSH_GATEWAY_URL"`
	CaCert                   string `json:"KAFKA_CA_CERT"`
	KafkaCaCertPath          string
	AwsConfig                *aws.Config
	KafkaBrokers             []string
}

var Cfg Config

func createTempCAFile(cert string) (string, error) {
	tmpPath := filepath.Join(os.TempDir(), "kafka-ca.pem")
	err := os.WriteFile(tmpPath, []byte(cert), 0644)
	if err != nil {
		return "", err
	}
	return tmpPath, nil
}

func Load(secretName string) {
	ctx := context.Background()

	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-2"))
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}

	smClient := secretsmanager.NewFromConfig(awsCfg)
	resp, err := smClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		log.Fatalf("unable to retrieve secrets: %v", err)
	}

	err = json.Unmarshal([]byte(*resp.SecretString), &Cfg)
	if err != nil {
		log.Fatalf("unable to unmarshal secrets: %v", err)
	}

	Cfg.AwsConfig = &awsCfg
	Cfg.KafkaBrokers = []string{}
	for _, broker := range strings.Split(Cfg.Brokers, ",") {
		Cfg.KafkaBrokers = append(Cfg.KafkaBrokers, strings.TrimSpace(broker))
	}

	if Cfg.KafkaCaCertPath == "" {
		Cfg.KafkaCaCertPath, err = createTempCAFile(Cfg.CaCert)
		if err != nil {
			log.Fatalf("unable to create temporary CA file: %v", err)
		}
	}
}
