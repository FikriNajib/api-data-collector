package v1

import (
	"data-collector-api/common"
	"data-collector-api/domain"
	"data-collector-api/service/behavior"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	"log"
)

type handler struct {
	service behavior.Behavior
}

func NewHandler(svc behavior.Behavior) *handler {
	return &handler{
		service: svc,
	}
}

func (h *handler) Handle(c *fiber.Ctx) error {
	span := sentry.StartSpan(c.Context(), "handler", sentry.TransactionName("API POST /api/v1/collector/behavior"))
	defer span.Finish()

	ctx := span.Context()

	userData := common.GetUserData(c)
	userId := int(userData["vid"].(float64))
	deviceId := userData["device_id"].(string)
	platform := userData["pl"].(string)
	userAgent := c.Get("User-Agent")
	var p domain.DataCollectorRequest
	fmt.Println("body req : ", string(c.Body()))
	if err := c.BodyParser(&p); err != nil {
		log.Println("error Body Parser", err)
		return c.Status(400).JSON(fiber.Map{
			"message": "Bad request",
		})
	}
	p.UserID = userId
	p.DeviceID = deviceId
	p.Platform = platform
	if p.UserID == 0 || p.UserID == 0.0 {
		p.UserID = p.DeviceID
	}
	p.UserAgent = userAgent

	if err := h.service.CreateBehavior(ctx, &p); err != nil {
		log.Println("Failed Create Behaviour :", err)
		sentry.CaptureException(err)
	}

	return c.Status(200).SendString("OK")
}
