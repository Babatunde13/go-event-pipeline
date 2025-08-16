package kafka

import (
	"context"
	"log"
	"time"

	"github.com/Babatunde13/event-pipeline/internal/config"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Producer struct {
	writer *kafka.Producer
}

type Consumer struct {
	reader *kafka.Consumer
}

func getKafkaConfig() (*kafka.ConfigMap, error) {
	cred, err := config.Cfg.AwsConfig.Credentials.Retrieve(context.Background())
	if err != nil {
		return nil, err
	}

	return &kafka.ConfigMap{
		"bootstrap.servers":           config.Cfg.Brokers,
		"security.protocol":           "SASL_SSL",
		"sasl.mechanism":              "AWS_MSK_IAM",
		"sasl.aws_msk_iam.access.key": cred.AccessKeyID,
		"sasl.aws_msk_iam.secret.key": cred.SecretAccessKey,
		"ssl.ca.pem":                  config.Cfg.CaCert,
	}, nil
}

func NewProducer() (*Producer, error) {
	kafkaConfig, err := getKafkaConfig()
	if err != nil {
		return nil, err
	}
	writer, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		return nil, err
	}

	return &Producer{
		writer: writer,
	}, nil
}

func CreateTopic(ctx context.Context, topicName string, numPartitions int, replicationFactor int) error {
	config, err := getKafkaConfig()
	if err != nil {
		return err
	}
	admin, err := kafka.NewAdminClient(config)
	if err != nil {
		return err
	}
	defer admin.Close()

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
	err := p.writer.Produce(&msg, deliveryChan)
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

func NewConsumer(brokers []string, topic string, groupID string) (*Consumer, error) {
	config, err := getKafkaConfig()
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
	reader, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, err
	}
	err = reader.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		reader.Close()
		return nil, err
	}
	log.Printf("Kafka consumer initialized with topic: %s", topic)
	return &Consumer{reader: reader}, nil
}

func (c *Consumer) ReadMessage(timeout time.Duration) (*kafka.Message, error) {
	return c.reader.ReadMessage(timeout)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
