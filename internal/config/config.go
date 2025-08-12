package config

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	secretsmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"log"
)

type Config struct {
	KafkaBrokers             []string `json:"KAFKA_BROKERS"`
	KafkaTopic               string   `json:"KAFKA_TOPIC"`
	EventBusName             string   `json:"EVENT_BUS_NAME"`
	RedisAddress             string   `json:"REDIS_ADDRESS"`
	PrometheusPort           string   `json:"PROMETHEUS_PORT"`
	Environment              string   `json:"ENVIRONMENT"`
	KafkaProducerPort        string   `json:"KAFKA_PRODUCER_PORT"`
	PrometheusPushGatewayUrl string   `json:"PROMETHEUS_PUSH_GATEWAY_URL"`
	AwsConfig                *aws.Config
}

var Cfg Config

func Load(secretName string) {
	ctx := context.Background()

	awsCfg, err := config.LoadDefaultConfig(ctx)
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

	if Cfg.KafkaProducerPort == "" {
		Cfg.KafkaProducerPort = "3000"
	}
	Cfg.AwsConfig = &awsCfg
}
