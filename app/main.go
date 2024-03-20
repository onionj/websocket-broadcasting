package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	WEB_SOCKET_HOST := os.Getenv("WEB_SOCKET_HOST")
	HOST := os.Getenv("HOST")
	RABBIT_URI := os.Getenv("RABBIT_URI")
	EXCHANGE_NAME := os.Getenv("EXCHANGE_NAME")

	app := fiber.New(fiber.Config{
		Prefork: true,
		Views:   html.New("./views", ".html"),
	})

	authConfig := basicauth.Config{
		Users: map[string]string{
			os.Getenv("PUSH_USERNAME"): os.Getenv("PUSH_PASSWORD"),
		},
	}

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())

	notifier := NewNotifier()

	// RabbitMQ
	conn, err := amqp.Dial(RABBIT_URI)
	failOnError("Failed to connect to RabbitMQ", err)
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError("Failed to open a channel", err)
	defer ch.Close()

	err = ch.ExchangeDeclare(EXCHANGE_NAME, "fanout", false, false, false, false, nil)
	failOnError("Failed to declare a Exchange", err)

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	failOnError("Failed to declare a queue", err)

	err = ch.QueueBind(q.Name, "", EXCHANGE_NAME, false, nil)
	failOnError("Failed to bind a queue", err)
	log.Printf("Binding queue %s to exchange %s with routing key", q.Name, EXCHANGE_NAME)

	// Subscribing to QueueService1 for getting messages.
	qMessages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	failOnError("Failed to Consume", err)

	go func() {
		for d := range qMessages {
			messageJSON, err := notifier.CreateMessage(string(d.Body))
			if err != nil {
				fmt.Println(err)
				continue
			}
			notifier.Push(messageJSON)
		}
	}()

	// WebSocket Upgrade Middleware
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// WebSocket Route Handler
	app.Get("/ws/notifier", websocket.New(func(conn *websocket.Conn) {
		notifier.AddConnection(conn)
		defer func() {
			notifier.RemoveConnection(conn)
			conn.Close()
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))

	// push notification
	app.Post("/push", basicauth.New(authConfig), func(c *fiber.Ctx) error {
		var requestBody struct {
			Content string `json:"content"`
		}
		if err := c.BodyParser(&requestBody); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = ch.PublishWithContext(ctx,
			EXCHANGE_NAME, // exchange
			"",            // routing key
			false,         // mandatory
			false,         // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(requestBody.Content),
			})
		log.Println("Failed to publish a message", err)

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"content": requestBody.Content})
	})

	// Render index template
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title":         "WebSocket Broadcasting",
			"WebSocketHost": WEB_SOCKET_HOST,
		})
	})

	log.Fatal(app.Listen(HOST))
}

func failOnError(msg string, err error) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
