package behavior

import (
	"context"
	"data-collector-api/app/infrastructure/kafka"
	"data-collector-api/config"
	"data-collector-api/domain"
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/getsentry/sentry-go"
	"log"
	"net/http"
	"time"
)

type Behavior interface {
	CreateBehavior(ctx context.Context, request *domain.DataCollectorRequest) error
}

type behaviorService struct {
	kp *kafka.Publisher
}

func NewBehaviorService(kp *kafka.Publisher) Behavior {
	return &behaviorService{
		kp: kp,
	}
}

func (s *behaviorService) CreateBehavior(ctx context.Context, request *domain.DataCollectorRequest) error {
	span := sentry.StartSpan(ctx, "handle")
	defer span.Finish()

	ctx = span.Context()

	s.addBehavior(ctx, request)

	return nil
}

// publish publishing the service
func (s *behaviorService) publish(ctx context.Context, topic string, request *domain.DataCollectorRequest) error {
	span := sentry.StartSpan(ctx, "publish")
	defer span.Finish()

	jsonString, err := json.Marshal(request)
	if err != nil {
		sentry.CaptureException(err)
		log.Println("Error Marshal: ", err)
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), jsonString)
	if err = s.kp.Publish(topic, msg); err != nil {
		sentry.CaptureException(err)
		log.Println("Error publish: ", err)
		return err
	}

	return nil
}

// addBehavior adding the behavior
func (s *behaviorService) addBehavior(ctx context.Context, p *domain.DataCollectorRequest) {
	span := sentry.StartSpan(ctx, "addBehavior")
	defer span.Finish()
	clientRealIP := ctx.Value("ClientRealIP")
	fmt.Println("ClientRealIP=>>", clientRealIP)

	ipaddress := s.CheckIP(clientRealIP.(string))
	p.IpAddress = ipaddress
	p.DateTime = time.Now()

	hystrix.ConfigureCommand("publish", hystrix.CommandConfig{
		Timeout:               config.Config.GetInt("HYSTRIX_TIMEOUT"),
		MaxConcurrentRequests: config.Config.GetInt("HYSTRIX_MAX_CONCURRENT_REQUESTS"),
		ErrorPercentThreshold: config.Config.GetInt("HYSTRIX_ERROR_PERCENT_THRESHOLD"),
	})

	hystrix.Go("publish", func() error {
		return s.publish(ctx, config.Config.GetString("TOPIC_BEHAVIOR_VIEW"), p)
	}, func(err error) error {
		sentry.CaptureException(err)

		log.Println("Circuit breaker opened for publish command", err.Error())

		return nil
	})

	return
}

func (s *behaviorService) CheckIP(ip string) string {
	resultChan := make(chan *string)

	// Configuring the Hystrix circuit
	hystrix.ConfigureCommand("checkIP", hystrix.CommandConfig{
		Timeout:               config.Config.GetInt("HYSTRIX_TIMEOUT"),
		MaxConcurrentRequests: config.Config.GetInt("HYSTRIX_MAX_CONCURRENT_REQUESTS"),
		ErrorPercentThreshold: config.Config.GetInt("HYSTRIX_ERROR_PERCENT_THRESHOLD"),
	})

	// Using hystrix.Do to execute the HTTP request with circuit breaker protection
	errChan := hystrix.Go("checkIP", func() error {
		ipResult := s.GetHttpCheckIP(ip)
		resultChan <- &ipResult
		return nil
	}, nil)
	select {
	case result := <-resultChan:
		log.Println("success:", result)
		return *result
	case errors := <-errChan:
		log.Println("Circuit breaker opened for http command", errors.Error())
		return ""
	}
	return ""
}

func (s *behaviorService) GetHttpCheckIP(ip string) string {
	var r map[string]interface{}

	url := "https://cek.rcti.plus/ip/" + ip
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return ""
	}

	res, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer res.Body.Close()

	if err2 := json.NewDecoder(res.Body).Decode(&r); err2 != nil {
		log.Fatalf("Error parsing the response body: %s", err2)
	}

	ipAddress, ok := r["ip_address"].(string)
	if !ok {
		fmt.Println("err get IP")
	}
	return ipAddress
}
