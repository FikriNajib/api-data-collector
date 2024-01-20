package config

import (
	"github.com/sahalazain/go-common/config"
)

var DefaultConfig = map[string]interface{}{
	"KAFKA_BROKER":                    "localhost:29092",
	"HYSTRIX_TIMEOUT":                 1000,
	"HYSTRIX_MAX_CONCURRENT_REQUESTS": 100,
	"HYSTRIX_ERROR_PERCENT_THRESHOLD": 25,
	"SENTRY_DSN":                      "",
	"SENTRY_TRACES_SAMPLE_RATE":       0.1,
	"SENTRY_ENVIRONMENT":              "development",
	"SENTRY_DEBUG":                    false,
	"KAFKA_CONSUMER_GROUP":            "user_behavior_group",
	"MONGO_URL":                       "mongodb://localhost:27017",
	"MONGO_DB":                        "userbehaviour",
	"MONGO_COLLECTION":                "behaviour",
	"TOPIC_BEHAVIOR_VIEW":             "user_behavior",
	"TOPIC_BEHAVIOR_VIEW_DLQ":         "local_userbehavior_dlq",
}

var Config config.Getter
var Url string

func Load() error {
	cfgClient, err := config.Load(DefaultConfig, Url)
	if err != nil {
		return err
	}

	Config = cfgClient

	return nil
}
