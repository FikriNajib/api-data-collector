package api

import (
	handlerV1 "data-collector-api/app/api/handler/behavior/v1"
	handlerV2 "data-collector-api/app/api/handler/behavior/v2"
	"data-collector-api/app/infrastructure/kafka"
	infrashared "data-collector-api/app/infrastructure/shared"
	"data-collector-api/config"
	"data-collector-api/service/behavior"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/contrib/fibersentry"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/golang-jwt/jwt/v4"
	"go.elastic.co/apm/module/apmfiber/v2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

var kp *kafka.Publisher

func configure() *fiber.App {
	kp = kafka.NewPublisher()

	// To initialize Sentry's handler, you need to initialize Sentry itself beforehand
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.Config.GetString("SENTRY_DSN"),
		TracesSampleRate: config.Config.GetFloat64("SENTRY_TRACES_SAMPLE_RATE"),
		Environment:      config.Config.GetString("SENTRY_ENVIRONMENT"),
		EnableTracing:    true,
		Debug:            config.Config.GetBool("SENTRY_DEBUG"),
		AttachStacktrace: true,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	app := fiber.New()
	app.Use(apmfiber.Middleware())
	app.Use(fibersentry.New(fibersentry.Config{
		Repanic:         true,
		WaitForDelivery: true,
	}))
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(pprof.New())

	return app
}

func Serve() {
	app := configure()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		_ = <-c
		fmt.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	app.Use(setBearerRule)
	app.Use(jwtware.New(jwtware.Config{
		SigningMethod: "HS256",
		SigningKey:    []byte("secret"),
	}))
	app.Use(validateJWTClient)
	app.Use(getClientRealIP)
	app = loadRoute(app)

	if err := app.Listen(":8011"); err != nil {
		log.Println("Failed to start server", err.Error())
	}

	fmt.Println("Running cleanup tasks...")
}

func loadRoute(app *fiber.App) *fiber.App {
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("Hello, World!") })
	app.Get("/ping", func(c *fiber.Ctx) error { return c.SendString("PONG!") })

	svcBehavior := behavior.NewBehaviorService(kp)

	v1 := handlerV1.NewHandler(svcBehavior)

	app.Post("/api/v1/collector/behavior", v1.Handle)

	return app
}

func setBearerRule(c *fiber.Ctx) error {
	tokenString := c.Get("Authorization")
	if strings.HasPrefix(tokenString, "Bearer") == false {
		c.Request().Header.Set("Authorization", "Bearer "+tokenString)
	}
	return c.Next()
}

func validateJWTClient(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	if claims, ok := user.Claims.(jwt.MapClaims); ok {
		c.Locals("user_data", map[string]interface{}(claims))
		return c.Next()
	}

	return c.SendStatus(http.StatusUnauthorized)
}

func getClientRealIP(c *fiber.Ctx) error {
	// Get the client's real IP address
	reqHeader := c.Request().Header
	ip := infrashared.GetClientRealIP(&reqHeader)

	// Set the client's real IP address in the request context
	c.Context().SetUserValue("ClientRealIP", ip)
	c.Context().SetUserValue("User-Agent", c.Request().Header.UserAgent())
	return c.Next()
}
