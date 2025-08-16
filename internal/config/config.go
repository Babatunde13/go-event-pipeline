package config

import (
	"context"
	"encoding/json"
	"log"
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

func parseCaCert(cert string) string {
	// Remove any leading or trailing whitespace
	cert = strings.TrimSpace(cert)
	// Ensure BEGIN and END markers are separate
	cert = strings.ReplaceAll(cert, "-----BEGIN CERTIFICATE----- ", "-----BEGIN CERTIFICATE-----\n")
	cert = strings.ReplaceAll(cert, " -----END CERTIFICATE-----", "\n-----END CERTIFICATE-----")

	// Extract the middle base64 part
	parts := strings.Split(cert, "\n")
	if len(parts) < 2 {
		log.Fatal("Invalid certificate format")
	}

	begin := parts[0]
	end := parts[len(parts)-1]

	// Split the base64 content by spaces → insert newlines every ~64 chars
	body := strings.ReplaceAll(strings.Join(parts[1:len(parts)-1], ""), " ", "")
	var chunks []string
	for len(body) > 64 {
		chunks = append(chunks, body[:64])
		body = body[64:]
	}
	if len(body) > 0 {
		chunks = append(chunks, body)
	}

	return begin + "\n" + strings.Join(chunks, "\n") + "\n" + end
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

	Cfg.CaCert = parseCaCert(Cfg.CaCert)
}
