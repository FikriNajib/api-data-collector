package kafka

import (
	"data-collector-api/config"
	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"log"
	"strings"
)

type Subscriber struct {
	*kafka.Subscriber
}

func NewSubscriber() *Subscriber {
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	saramaSubscriberConfig := kafka.DefaultSaramaSubscriberConfig()
	// equivalent of auto.offset.reset: earliest
	saramaSubscriberConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	subscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               strings.Split(config.Config.GetString("KAFKA_BROKER"), ","),
			Unmarshaler:           kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: saramaSubscriberConfig,
			ConsumerGroup:         config.Config.GetString("KAFKA_CONSUMER_GROUP"),
		},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		panic(err)
	}

	return &Subscriber{subscriber}
}
