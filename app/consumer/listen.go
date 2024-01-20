package consumer

import (
	"context"
	"data-collector-api/app/consumer/handler"
	"data-collector-api/app/infrastructure"
	kafka2 "data-collector-api/app/infrastructure/kafka"
	"data-collector-api/config"
	"data-collector-api/service/behavior"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	"time"
)

var (
	logger = watermill.NewStdLogger(false, false)
)

func configure() *message.Router {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}

	router.AddPlugin(plugin.SignalsHandler)

	// Router level middleware are executed for every message sent to the router
	router.AddMiddleware(
		middleware.CorrelationID,

		middleware.Retry{
			MaxRetries:      3,
			InitialInterval: time.Millisecond * 100,
			Logger:          logger,
		}.Middleware,

		middleware.Recoverer,
	)

	return router
}

func Listen() {
	publisher := kafka2.NewPublisher()
	subscriber := kafka2.NewSubscriber()

	router := configure()
	mongo := infrastructure.AppMongo
	svcConsumerBehavior := behavior.NewConsumerBehaviorService(subscriber, mongo)
	behaviourHandler := handler.NewConsumerBehaviorHandler(svcConsumerBehavior)

	router.AddHandler(
		"user_behavior",                           // handler name, must be unique
		config.Config.GetString("TOPIC_BEHAVIOR_VIEW"), // topic from which we will read events
		subscriber,
		config.Config.GetString("TOPIC_BEHAVIOR_VIEW_DLQ"), // topic to which we will publish events
		publisher,
		behaviourHandler.BehaviourHandle,
	)

	ctx := context.Background()
	if err := router.Run(ctx); err != nil {
		panic(err)
	}
}
