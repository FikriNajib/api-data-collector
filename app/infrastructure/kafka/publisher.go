package kafka

import (
	"data-collector-api/config"
	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"strings"
)

type Publisher struct {
	*kafka.Publisher
}

func NewPublisher() *Publisher {
	saramaSubscriberConfig := kafka.DefaultSaramaSubscriberConfig()
	saramaSubscriberConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	publisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   strings.Split(config.Config.GetString("KAFKA_BROKER"), ","),
			Marshaler: kafka.DefaultMarshaler{},
		},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		panic(err)
	}

	return &Publisher{publisher}
}
