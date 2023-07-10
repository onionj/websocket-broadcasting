package main

import (
	"log"
	"os"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	WEB_SOCKET_HOST := os.Getenv("WEB_SOCKET_HOST")
	HOST := os.Getenv("HOST")

	// Basic Auth Middleware
	authConfig := basicauth.Config{
		Users: map[string]string{
			os.Getenv("PUSH_USERNAME"): os.Getenv("PUSH_PASSWORD"),
		},
	}

	htmlEngine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: htmlEngine,
	})

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

	app.Post("/push", basicauth.New(authConfig), func(c *fiber.Ctx) error {
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

	app.Get("/", func(c *fiber.Ctx) error {
		// Render index template
		return c.Render("index", fiber.Map{
			"Title":         "WebSocket Broadcasting",
			"WebSocketHost": WEB_SOCKET_HOST,
		})
	})

	log.Fatal(app.Listen(HOST))
}
