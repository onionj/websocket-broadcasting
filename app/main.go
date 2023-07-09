package main

import (
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	app := fiber.New()
	notifier := NewNotifier()

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())

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

	// Push Notification Endpoint
	app.Post("/push", func(c *fiber.Ctx) error {
		var requestBody struct {
			Content string `json:"content"`
		}
		if err := c.BodyParser(&requestBody); err != nil {
			return err
		}

		messageJSON, err := notifier.CreateMessage(requestBody.Content)
		if err != nil {
			return err
		}

		notifier.Push(messageJSON)

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"content": requestBody.Content})
	})

	log.Fatal(app.Listen(":3000"))
}
