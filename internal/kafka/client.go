package kafka

import (
	"context"
	"log"
	"time"

	"github.com/Babatunde13/event-pipeline/internal/config"
	"github.com/aws/aws-msk-iam-sasl-signer-go/signer"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Producer struct {
	producer *kafka.Producer
}

type Consumer struct {
	consumer *kafka.Consumer
}

func createTokenProvider() (*kafka.OAuthBearerToken, error) {
	token, tokenExpirationTime, err := signer.GenerateAuthToken(context.TODO(), "us-east-2")
	if err != nil {
		return nil, err
	}
	seconds := tokenExpirationTime / 1000
	nanoseconds := (tokenExpirationTime % 1000) * 1000000
	bearerToken := kafka.OAuthBearerToken{
		TokenValue: token,
		Expiration: time.Unix(seconds, nanoseconds),
	}
	return &bearerToken, nil
}
func getKafkaConfig() (*kafka.ConfigMap, *kafka.OAuthBearerToken, error) {
	tokenProvider, err := createTokenProvider()
	if err != nil {
		return nil, nil, err
	}
	return &kafka.ConfigMap{
		"bootstrap.servers": config.Cfg.Brokers,
		"security.protocol": "SASL_SSL",
		"sasl.mechanisms":   "OAUTHBEARER",
		"ssl.ca.pem":        config.Cfg.CaCert,
	}, tokenProvider, nil
}

func NewProducer() (*Producer, error) {
	kafkaConfig, tokenProvider, err := getKafkaConfig()
	if err != nil {
		return nil, err
	}
	producer, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		return nil, err
	}
	producer.SetOAuthBearerToken(*tokenProvider)
	return &Producer{
		producer: producer,
	}, nil
}

func CreateTopic(ctx context.Context, topicName string, numPartitions int, replicationFactor int) error {
	config, tokenProvider, err := getKafkaConfig()
	if err != nil {
		return err
	}
	admin, err := kafka.NewAdminClient(config)
	if err != nil {
		return err
	}
	defer admin.Close()
	admin.SetOAuthBearerToken(*tokenProvider)

	// Define topic spec
	topic := kafka.TopicSpecification{
		Topic:             topicName,
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
	}
	// Create topic
	results, err := admin.CreateTopics(ctx, []kafka.TopicSpecification{topic}, kafka.SetAdminOperationTimeout(10*time.Second))
	if err != nil {
		return err
	}

	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError {
			log.Printf("Failed to create topic %s: %v", result.Topic, result.Error)
			return result.Error
		}
		log.Printf("Topic %s created successfully", result.Topic)
	}
	return nil
}

func (p *Producer) SendMessage(ctx context.Context, key string, value []byte) error {
	msg := kafka.Message{
		Key:   []byte(key),
		Value: value,
		TopicPartition: kafka.TopicPartition{
			Topic:     &config.Cfg.KafkaTopic,
			Partition: kafka.PartitionAny,
		},
	}
	deliveryChan := make(chan kafka.Event, 1)
	err := p.producer.Produce(&msg, deliveryChan)
	if err != nil {
		return err
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		return m.TopicPartition.Error
	}
	return nil
}

func NewConsumer(topic string, groupID string) (*Consumer, error) {
	config, tokenProvider, err := getKafkaConfig()
	if err != nil {
		return nil, err
	}

	config.SetKey("group.id", groupID)
	config.SetKey("auto.offset.reset", "earliest")
	config.SetKey("enable.auto.commit", "true")
	config.SetKey("session.timeout.ms", 6000)
	config.SetKey("max.poll.interval.ms", 300000) // 5 minutes
	config.SetKey("fetch.min.bytes", 10000)       // 10KB
	config.SetKey("fetch.max.bytes", 10000000)    // 10MB
	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, err
	}
	consumer.SetOAuthBearerToken(*tokenProvider)
	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		consumer.Close()
		return nil, err
	}
	log.Printf("Kafka consumer initialized with topic: %s", topic)
	return &Consumer{consumer: consumer}, nil
}

func (c *Consumer) ReadMessage(timeout time.Duration) (*kafka.Message, error) {
	return c.consumer.ReadMessage(timeout)
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}
