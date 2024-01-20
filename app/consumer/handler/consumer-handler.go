package handler

import (
	"context"
	"data-collector-api/app/infrastructure"
	"data-collector-api/app/infrastructure/kafka"
	"data-collector-api/domain"
	"data-collector-api/service/behavior"
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/getsentry/sentry-go"
	"log"
)

var (
	kp *kafka.Subscriber
	m  *infrastructure.DBMongo
)

type ConsumerBehaviorHandler struct {
	service behavior.ConsumerBehavior
}

func NewConsumerBehaviorHandler(svc behavior.ConsumerBehavior) *ConsumerBehaviorHandler {
	return &ConsumerBehaviorHandler{service: svc}
}

func (s ConsumerBehaviorHandler) BehaviourHandle(msg *message.Message) ([]*message.Message, error) {
	fmt.Printf(
		"\n> Received message: %s\n> %s\n> metadata: %v\n\n",
		msg.UUID, string(msg.Payload), msg.Metadata,
	)
	var behaviour domain.DataCollectorRequest
	ctx := context.Background()

	if err := json.Unmarshal(msg.Payload, &behaviour); err != nil {
		log.Println("Error unmarshal", err)
		sentry.CaptureException(err)
	}
	if err := s.service.ProcessMessage(ctx, &behaviour); err != nil {
		log.Println("Failed Consume/Insert Mongo", err)
		sentry.CaptureException(err)
	}
	log.Println("Success insert mongo")
	return message.Messages{msg}, nil
}
