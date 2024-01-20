package behavior

import (
	"context"
	"data-collector-api/app/infrastructure"
	"data-collector-api/app/infrastructure/kafka"
	"data-collector-api/domain"
	"data-collector-api/domain/entities"
	"github.com/getsentry/sentry-go"
	"time"
)

type ConsumerBehavior interface {
	ProcessMessage(ctx context.Context, behaviour *domain.DataCollectorRequest) error
}

type consumerBehaviorService struct {
	kp    *kafka.Subscriber
	mongo *infrastructure.DBMongo
}

func NewConsumerBehaviorService(kp *kafka.Subscriber, m *infrastructure.DBMongo) ConsumerBehavior {
	return &consumerBehaviorService{
		kp:    kp,
		mongo: m,
	}
}

func (c consumerBehaviorService) ExistingUserIdFound(ctx context.Context, behaviour *domain.DataCollectorRequest) (bool, error) {
	span := sentry.StartSpan(ctx, "existingUserIdFound")
	defer span.Finish()

	ctx = span.Context()

	if _, err := c.mongo.GetDataByUserID(ctx, behaviour.UserID); err != nil {
		if err.Error() == "mongo: no documents in result" {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (c consumerBehaviorService) UpdateRecord(ctx context.Context, behaviour *domain.DataCollectorRequest) error {
	span := sentry.StartSpan(ctx, "updateRecord")
	defer span.Finish()

	ctx = span.Context()

	behaviourData := entities.Action{
		Service:          behaviour.Service,
		ContentType:      behaviour.ContentType,
		ContentID:        behaviour.ContentID,
		ActionUser:       behaviour.Action,
		Title:            behaviour.Title,
		Name:             behaviour.Name,
		DeviceID:         behaviour.DeviceID,
		Platform:         behaviour.Platform,
		DateTime:         time.Now(),
		Hashtag:          behaviour.Hashtag,
		Pillar:           behaviour.Pillar,
		Duration:         behaviour.Duration,
		IpAddress:        behaviour.IpAddress,
		CreatorID:        behaviour.CreatorID,
		VideoDuration:    behaviour.VideoDuration,
		IsIncognito:      behaviour.IsIncognito,
		IsEmbeddedIframe: behaviour.IsEmbeddedIframe,
		UserAgent:        behaviour.UserAgent,
	}

	return c.mongo.UpdatePushDetail(ctx, behaviourData, behaviour.UserID, "action")
}

func (c consumerBehaviorService) ProcessMessage(ctx context.Context, behaviour *domain.DataCollectorRequest) error {
	span := sentry.StartSpan(ctx, "processMessage", sentry.TransactionName("consumer.behaviour.view"))
	defer span.Finish()

	ctx = span.Context()

	activity := entities.UserActivity{
		UserId: behaviour.UserID,
		Status: "init",
		Action: []entities.Action{{
			Service:          behaviour.Service,
			ContentType:      behaviour.ContentType,
			ContentID:        behaviour.ContentID,
			ActionUser:       behaviour.Action,
			Title:            behaviour.Title,
			Name:             behaviour.Name,
			DeviceID:         behaviour.DeviceID,
			Platform:         behaviour.Platform,
			DateTime:         time.Now(),
			Hashtag:          behaviour.Hashtag,
			Pillar:           behaviour.Pillar,
			Duration:         behaviour.Duration,
			IpAddress:        behaviour.IpAddress,
			CreatorID:        behaviour.CreatorID,
			VideoDuration:    behaviour.VideoDuration,
			IsIncognito:      behaviour.IsIncognito,
			IsEmbeddedIframe: behaviour.IsEmbeddedIframe,
			UserAgent:        behaviour.UserAgent,
		}},
	}

	exist, err := c.ExistingUserIdFound(ctx, behaviour)
	if err != nil {
		sentry.CaptureException(err)

		return err
	}

	if exist {
		if err = c.UpdateRecord(ctx, behaviour); err != nil {
			sentry.CaptureException(err)

			return nil
		}

		return nil
	}

	return c.mongo.InsertMongo(ctx, activity)
}
